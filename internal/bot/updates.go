package bot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"tg_bot_asist/internal/credits"
	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/storage"
	"tg_bot_asist/internal/todo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleUpdates обрабатывает входящие обновления от Telegram API.
// Функция запускает основной цикл обработки сообщений с автоматической перезагрузкой
// при ошибках Telegram API (409, 500) и graceful shutdown при получении сигналов.
// DEPRECATED: Используйте HandleUpdatesWithContext для поддержки контекста.
func HandleUpdates(
	bot *tgbotapi.BotAPI,
	updates tgbotapi.UpdatesChannel,
	userRepo *storage.UserRepo,
	todoService *todo.Service,
	stateRepo *storage.StateRepo,
	creditService *credits.Service,
	financeService *finance.Service,
	recurringRepo *finance.RecurringRepo,
) {
	ctx := context.Background()
	HandleUpdatesWithContext(ctx, bot, updates, userRepo, todoService, stateRepo, creditService, financeService, recurringRepo)
}

// HandleUpdatesWithContext обрабатывает входящие обновления от Telegram API с поддержкой контекста.
// Функция запускает основной цикл обработки сообщений с автоматической перезагрузкой
// при ошибках Telegram API (409, 500) и graceful shutdown при отмене контекста.
func HandleUpdatesWithContext(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	updates tgbotapi.UpdatesChannel,
	userRepo *storage.UserRepo,
	todoService *todo.Service,
	stateRepo *storage.StateRepo,
	creditService *credits.Service,
	financeService *finance.Service,
	recurringRepo *finance.RecurringRepo,
) {
	// Настройка Menu Button для WebApp (если нужно, настройте через BotFather или используйте команду)
	// Для настройки через код нужна поддержка в библиотеке telegram-bot-api
	// Пока настраивается вручную через BotFather: /mybots → Bot Settings → Menu Button
	// Создаём простой FSM (in-memory) и Handler
	fsm := NewFSM()
	handler := NewHandler(bot, fsm, todoService, financeService, creditService)

	// Основной цикл обработки обновлений
	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down bot...")
			return

		case update, ok := <-updates:
			if !ok {
				// Канал закрыт - возможен конфликт с другим экземпляром
				logger.Warn("Updates channel closed, possible conflict detected")

				// Останавливаем возможные конфликтующие процессы
				killConflictingProcesses()

				// Ждём дольше, чтобы Telegram API освободил соединение
				logger.Info("Waiting 5 seconds before reconnecting...")
				time.Sleep(5 * time.Second)

				// Пытаемся переподключиться
				u := tgbotapi.NewUpdate(0)
				u.Timeout = 60
				updates = bot.GetUpdatesChan(u)
				logger.Info("Reconnected to Telegram API")
				continue
			}

			// Обрабатываем обновление в отдельной горутине для безопасности
			u := update
			go func(update tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("Panic in update handler: " + string(r.(string)))
					}
				}()

				// Регистрируем пользователя при первом обращении
				if u.Message != nil && u.Message.From != nil {
					userCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					chatID := u.Message.Chat.ID
					lastMessage := ""
					if u.Message.Text != "" {
						lastMessage = u.Message.Text
					}
					if err := userRepo.RegisterUser(userCtx, u.Message.From.ID, chatID, lastMessage); err != nil {
						logger.Error("Failed to register user: " + err.Error())
					}
				}

				// Обрабатываем сообщение
				handler.Handle(update)
			}(u)
		}
	}
}

// killConflictingProcesses останавливает возможные конфликтующие процессы бота и зомби-процессы.
// Вызывается при обнаружении конфликта getUpdates.
// ВАЖНО: Не останавливает текущий процесс (использует PID для исключения).
func killConflictingProcesses() {
	logger.Info("Stopping conflicting bot processes and zombies...")

	currentPID := os.Getpid()

	var commands []*exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows: останавливаем все процессы, кроме текущего
		commands = []*exec.Cmd{
			exec.Command("taskkill", "/F", "/IM", "go.exe", "/FI", "PID ne "+fmt.Sprintf("%d", currentPID)),
			exec.Command("taskkill", "/F", "/IM", "tg_bot_asist.exe", "/FI", "PID ne "+fmt.Sprintf("%d", currentPID)),
		}
	} else {
		// Unix-like: используем pgrep для поиска процессов, исключая текущий PID
		// Останавливаем процессы, которые НЕ являются текущим процессом
		script := fmt.Sprintf(`
			# Останавливаем обычные процессы
			pgrep -f "go run.*main.go" | grep -v "^%d$" | xargs kill -TERM 2>/dev/null || true
			pgrep -f "tg_bot_asist" | grep -v "^%d$" | xargs kill -TERM 2>/dev/null || true
			
			# Останавливаем процессы на порту 8080 (зомби API серверы)
			lsof -ti:8080 2>/dev/null | grep -v "^%d$" | xargs kill -TERM 2>/dev/null || true
		`, currentPID, currentPID, currentPID)

		cmd1 := exec.Command("sh", "-c", script)
		cmd1.Stdout = os.Stdout
		cmd1.Stderr = os.Stderr
		commands = append(commands, cmd1)
	}

	// Выполняем команды
	for _, cmd := range commands {
		_ = cmd.Run() // Игнорируем ошибки
	}

	// Даём время процессам завершиться
	time.Sleep(2 * time.Second)

	// Если процессы всё ещё есть, используем более жёсткий метод (но не для текущего PID)
	if runtime.GOOS != "windows" {
		script := fmt.Sprintf(`
			# Жёсткая остановка всех процессов, включая зомби
			pgrep -f "go run.*main.go" | grep -v "^%d$" | xargs kill -9 2>/dev/null || true
			pgrep -f "tg_bot_asist" | grep -v "^%d$" | xargs kill -9 2>/dev/null || true
			lsof -ti:8080 2>/dev/null | grep -v "^%d$" | xargs kill -9 2>/dev/null || true
			
			# Очистка зомби-процессов (процессы, которые могут быть связаны с ботом)
			ps aux | grep -E "[g]o.*main.go|[g]o.*tg_bot" | grep -v "^.*%d " | \
			awk '{if ($2 != "%d") print $2}' | xargs kill -9 2>/dev/null || true
		`, currentPID, currentPID, currentPID, currentPID, currentPID)

		cmd2 := exec.Command("sh", "-c", script)
		_ = cmd2.Run()
		time.Sleep(1 * time.Second)
	}

	logger.Info("Conflicting processes and zombies cleanup completed")
}
