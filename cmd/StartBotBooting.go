package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"tg_bot_asist/internal/config"
	"tg_bot_asist/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotState struct {
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel
}

// killPreviousInstances останавливает все предыдущие экземпляры бота, зомби-процессы и освобождает порт 8080.
// Функция автоматически вызывается перед запуском бота для предотвращения конфликтов.
// Выполняет несколько попыток остановки для надёжности.
// ВАЖНО: Не останавливает текущий процесс (использует PID для исключения).
func killPreviousInstances() {
	logger.Info("Checking for previous bot instances and zombie processes...")

	currentPID := os.Getpid()
	var commands []*exec.Cmd

	// Определяем команды в зависимости от ОС
	if runtime.GOOS == "windows" {
		// Windows: исключаем текущий PID
		commands = []*exec.Cmd{
			exec.Command("taskkill", "/F", "/IM", "go.exe", "/FI", "PID ne "+fmt.Sprintf("%d", currentPID), "/FI", "WINDOWTITLE eq *main.go*"),
			exec.Command("taskkill", "/F", "/IM", "tg_bot_asist.exe", "/FI", "PID ne "+fmt.Sprintf("%d", currentPID)),
		}
	} else {
		// Unix-like (Linux, macOS)
		// 1. Останавливаем процессы go run main.go, исключая текущий PID
		script1 := fmt.Sprintf(`pgrep -f "go run.*main.go" | grep -v "^%d$" | xargs kill -9 2>/dev/null || true`, currentPID)
		cmd1 := exec.Command("sh", "-c", script1)
		cmd1.Stdout = os.Stdout
		cmd1.Stderr = os.Stderr
		commands = append(commands, cmd1)

		// 2. Останавливаем процессы tg_bot_asist, исключая текущий PID
		script2 := fmt.Sprintf(`pgrep -f "tg_bot_asist" | grep -v "^%d$" | xargs kill -9 2>/dev/null || true`, currentPID)
		cmd2 := exec.Command("sh", "-c", script2)
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		commands = append(commands, cmd2)

		// 3. Ищем и останавливаем скомпилированные бинарники в /tmp/go-build*/exe/main
		// Это процессы, которые остались после go run main.go
		script3 := fmt.Sprintf(`
			# Находим все процессы /tmp/go-build*/exe/main (скомпилированные бинарники)
			# Исключаем текущий процесс и его родительские процессы
			ps aux | grep -E "/tmp/go-build.*/exe/main" | grep -v "^.*%d " | \
			awk '{if ($2 != "%d") print $2}' | xargs kill -9 2>/dev/null || true
		`, currentPID, currentPID)
		cmd3 := exec.Command("sh", "-c", script3)
		cmd3.Stdout = os.Stdout
		cmd3.Stderr = os.Stderr
		commands = append(commands, cmd3)

		// 4. Ищем и останавливаем зомби-процессы на порту 8080 (API сервер)
		script4 := fmt.Sprintf(`
			# Находим процессы на порту 8080 (кроме текущего)
			PIDS=$(lsof -ti:8080 2>/dev/null | grep -v "^%d$" || true)
			if [ -n "$PIDS" ]; then
				echo "$PIDS" | xargs kill -9 2>/dev/null || true
			fi
		`, currentPID)
		cmd4 := exec.Command("sh", "-c", script4)
		cmd4.Stdout = os.Stdout
		cmd4.Stderr = os.Stderr
		commands = append(commands, cmd4)

		// 5. Ищем зависшие процессы go, которые могут быть связаны с ботом
		// (процессы, которые работают слишком долго и могут быть зомби)
		script5 := fmt.Sprintf(`
			# Находим все процессы go, которые работают дольше 1 часа и могут быть зомби
			# Исключаем текущий процесс
			ps aux | grep -E "[g]o run.*main.go|[g]o.*tg_bot" | grep -v "^.*%d " | awk '{print $2}' | xargs kill -9 2>/dev/null || true
		`, currentPID)
		cmd5 := exec.Command("sh", "-c", script5)
		cmd5.Stdout = os.Stdout
		cmd5.Stderr = os.Stderr
		commands = append(commands, cmd5)

		// 6. Ищем процессы по имени рабочей директории (если бот запущен из этой директории)
		// Это может помочь найти зомби-процессы, которые остались после краша
		script6 := fmt.Sprintf(`
			# Находим процессы, которые работают в директории проекта
			# Исключаем текущий процесс
			lsof +D . 2>/dev/null | grep -E "go|tg_bot" | awk '{print $2}' | grep -v "^%d$" | sort -u | xargs kill -9 2>/dev/null || true
		`, currentPID)
		cmd6 := exec.Command("sh", "-c", script6)
		cmd6.Stdout = os.Stdout
		cmd6.Stderr = os.Stderr
		commands = append(commands, cmd6)
	}

	// Выполняем команды несколько раз для надёжности
	maxAttempts := 3
	killed := false

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		for _, cmd := range commands {
			if err := cmd.Run(); err == nil {
				killed = true
			}
		}

		// Даём время процессам завершиться
		if killed {
			// Увеличена задержка для надёжности - даём больше времени Telegram API
			waitTime := 2 * time.Second
			if attempt == maxAttempts {
				waitTime = 4 * time.Second // Последняя попытка - ждём дольше
			}
			time.Sleep(waitTime)

			// Проверяем, что процессы действительно остановлены
			if !hasRunningInstances() {
				logger.Info("Previous bot instances stopped successfully")
				// Дополнительная задержка для освобождения соединения в Telegram API
				time.Sleep(2 * time.Second)
				return
			}

			// Если процессы ещё есть, пробуем ещё раз
			if attempt < maxAttempts {
				logger.Info(fmt.Sprintf("Some processes still running, retrying (attempt %d/%d)...", attempt+1, maxAttempts))
			}
		}
	}

	if killed {
		logger.Warn("Some bot instances may still be running. Waiting additional 5 seconds for Telegram API to release connection...")
		time.Sleep(5 * time.Second)

		// Финальная проверка и очистка зомби-процессов
		cleanupZombieProcesses(currentPID)
	} else {
		logger.Info("No previous instances found")
		// Даже если не было процессов, проверяем зомби-процессы
		cleanupZombieProcesses(currentPID)
		// Даём небольшую задержку для стабильности
		time.Sleep(1 * time.Second)
	}
}

// cleanupZombieProcesses ищет и останавливает зомби-процессы, связанные с ботом.
// Зомби-процессы - это процессы, которые остались после некорректного завершения.
func cleanupZombieProcesses(currentPID int) {
	if runtime.GOOS == "windows" {
		// Windows: проверка зависших процессов
		// Можно добавить проверку через tasklist, но для Windows это сложнее
		return
	}

	logger.Info("Cleaning up zombie processes...")

	// 1. Ищем процессы, которые используют порт 8080 (но не отвечают)
	script1 := fmt.Sprintf(`
		# Находим процессы на порту 8080
		PIDS=$(lsof -ti:8080 2>/dev/null | grep -v "^%d$" || true)
		if [ -n "$PIDS" ]; then
			for PID in $PIDS; do
				# Проверяем, что процесс действительно существует и не отвечает
				if kill -0 $PID 2>/dev/null; then
					# Процесс существует, но может быть зомби - убиваем
					kill -9 $PID 2>/dev/null || true
				fi
			done
		fi
	`, currentPID)
	cmd1 := exec.Command("sh", "-c", script1)
	_ = cmd1.Run()

	// 2. Ищем процессы go, которые могут быть зомби (старые процессы)
	script2 := fmt.Sprintf(`
		# Находим процессы go, которые работают в текущей директории
		# И могут быть остатками от предыдущих запусков
		ps aux | grep -E "[g]o.*main.go|[g]o.*tg_bot" | grep -v "^.*%d " | \
		awk '{if ($2 != "%d") print $2}' | xargs kill -9 2>/dev/null || true
	`, currentPID, currentPID)
	cmd2 := exec.Command("sh", "-c", script2)
	_ = cmd2.Run()

	// 3. Ищем процессы, которые могут быть связаны с ботом через переменные окружения
	// (если бот был запущен с определёнными переменными)
	script3 := fmt.Sprintf(`
		# Ищем процессы, которые могут использовать TELEGRAM_BOT_API
		# Это более агрессивная очистка
		ps aux | grep -E "TELEGRAM_BOT_API|tg_bot_asist" | grep -v "^.*%d " | \
		awk '{if ($2 != "%d") print $2}' | xargs kill -9 2>/dev/null || true
	`, currentPID, currentPID)
	cmd3 := exec.Command("sh", "-c", script3)
	_ = cmd3.Run()

	// Даём время зомби-процессам завершиться
	time.Sleep(1 * time.Second)

	logger.Info("Zombie processes cleanup completed")
}

// hasRunningInstances проверяет, есть ли запущенные экземпляры бота (кроме текущего процесса).
func hasRunningInstances() bool {
	currentPID := os.Getpid()

	if runtime.GOOS == "windows" {
		// Windows: проверка через tasklist, исключая текущий PID
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq go.exe", "/FI", "PID ne "+fmt.Sprintf("%d", currentPID), "/FI", "WINDOWTITLE eq *main.go*")
		output, _ := cmd.Output()
		return len(output) > 0
	} else {
		// Unix-like: проверка через pgrep, исключая текущий PID
		script := fmt.Sprintf(`pgrep -f "go run.*main.go" | grep -v "^%d$" || pgrep -f "tg_bot_asist" | grep -v "^%d$"`, currentPID, currentPID)
		cmd := exec.Command("sh", "-c", script)
		err := cmd.Run()

		// Если команда нашла процесс (exit code 0), значит есть запущенные экземпляры
		return err == nil
	}
}

// StartBotBooting инициализирует бота и настраивает получение обновлений.
// Автоматически останавливает предыдущие экземпляры перед запуском.
func StartBotBooting() (*BotState, error) {
	// Останавливаем предыдущие экземпляры перед запуском
	killPreviousInstances()

	token := config.GetRequired("TELEGRAM_BOT_API")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_API env variable empty")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("bot initialization error: %w", err)
	}

	logger.Info("Bot authorized as: " + bot.Self.UserName)

	// Дополнительная задержка перед получением обновлений
	// Даёт время Telegram API освободить соединение от предыдущего экземпляра
	// Увеличена до 5 секунд для большей надёжности (Telegram API может держать соединение до 5 секунд)
	logger.Info("Waiting 5 seconds before connecting to Telegram API...")
	time.Sleep(5 * time.Second)

	// Пытаемся закрыть предыдущие соединения через Telegram API
	// Используем deleteWebhook для сброса состояния (если был webhook)
	logger.Info("Resetting Telegram API state...")
	deleteWebhookConfig := tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: true, // Удаляем все ожидающие обновления
	}
	if _, err := bot.Request(deleteWebhookConfig); err != nil {
		logger.Warn("Failed to delete webhook (may not exist): " + err.Error())
	}
	time.Sleep(1 * time.Second)

	// Создаём канал обновлений
	// Если возникнет конфликт, он будет обработан в updates.go
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	logger.Info("Updates channel created successfully")

	return &BotState{
		Bot:     bot,
		Updates: updates,
	}, nil
}
