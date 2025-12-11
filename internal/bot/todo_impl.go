package bot

import (
	"context"
	"fmt"
	"strings"

	"tg_bot_asist/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// startTodoAdd запускает FSM для добавления новой задачи.
func (h *Handler) startTodoAdd(userID int64) {
	h.fsm.Set(userID, "TODO_ADD", map[string]any{})
	h.Send(userID, "Введи текст задачи:", BackKeyboard())
}

// handleTodoAdd: обрабатывает ввод текста задачи, сохраняет и закрывает FSM.
// Вызывается из Handler.handle (FSM) когда state.Name == "TODO_ADD".
func (h *Handler) handleTodoAdd(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	// проверяем, что у пользователя действительно есть FSM-стейт TODO_ADD
	state := h.fsm.Get(userID)
	if state == nil || state.Name != "TODO_ADD" {
		// нечего делать
		h.Send(chatID, "Нет активного создания задачи. Нажми «Добавить задачу»", TodoKeyboard())
		return
	}

	// простая логика: текст сообщения — это заголовок задачи
	title := text
	if title == "" {
		h.Send(chatID, "Пустая задача — введите текст задачи ещё раз:", BackKeyboard())
		return
	}

	// сохраняем через сервис
	if err := h.todo.Add(context.Background(), userID, title, "", nil); err != nil {
		logger.Error("Todo add failed: " + err.Error())
		h.Send(chatID, "Ошибка при сохранении задачи. Попробуйте позже.", BackKeyboard())
		return
	}

	// уведомляем пользователя и очищаем FSM
	h.Send(chatID, "Задача сохранена: "+title, HomeKeyboard())
	h.fsm.Clear(userID)
}

// showTodoList: выводит список задач для пользователя.
// Вызывается из Handler.showTodoList (кнопка / меню).
func (h *Handler) showTodoList(userID int64) {
	chatID := userID // в нашей архитектуре chatID == userID (private chat)
	ctx := context.Background()

	items, err := h.todo.List(ctx, userID)
	if err != nil {
		logger.Error("Todo list error: " + err.Error())
		h.Send(chatID, "Ошибка получения списка задач", TodoKeyboard())
		return
	}

	if len(items) == 0 {
		h.Send(chatID, "Список дел пуст", TodoKeyboard())
		return
	}

	// Формируем читаемое сообщение
	var b strings.Builder
	b.WriteString("Ваши задачи:\n\n")
	for _, it := range items {
		// ID и заголовок, можно добавить статус/дедлайн
		line := fmt.Sprintf("%d) %s\n", it.ID, it.Title)
		b.WriteString(line)
	}
	b.WriteString("\nЧтобы удалить задачу — отправь команду /delete_todo <id>")

	h.Send(chatID, b.String(), TodoKeyboard())
}
