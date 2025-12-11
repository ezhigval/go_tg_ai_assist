package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// -----------------------------
// FSM / Recurring implementation
// -----------------------------

// startRecurringAdd — вызывается из Handler.startRecurringAdd (помеченного ранее)
func (h *Handler) startRecurringAdd(userID int64) {
	// подготовим пустой черновик
	draft := map[string]any{
		"title":       "",
		"amount":      0.0,
		"period":      "",
		"payment_day": 0, // для monthly
	}
	h.fsm.Set(userID, "RECURRING_ADD", draft)

	// Просим название
	h.Send(userID, "Введите название регулярного платежа:", RecurringKeyboard())
}

// handleRecurringAdd — основной FSM-обработчик (вызывается из Handler.handle когда state active)
func (h *Handler) handleRecurringAdd(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	state := h.fsm.Get(userID)
	if state == nil || state.Name != "RECURRING_ADD" {
		h.Send(chatID, "Нет активного добавления регулярки.", RecurringKeyboard())
		return
	}

	// draft у нас map[string]any, используем напрямую
	draft := state.Data
	if draft == nil {
		// защита: восстановим структуру
		draft = map[string]any{
			"title":       "",
			"amount":      0.0,
			"period":      "",
			"payment_day": 0,
		}
		h.fsm.Set(userID, "RECURRING_ADD", draft)
	}

	// определим текущий шаг по заполненным полям
	// шаги: title -> amount -> period -> (payment_day if monthly) -> confirm (auto-save)
	if draft["title"] == "" {
		draft["title"] = text
		h.fsm.Set(userID, "RECURRING_ADD", draft)
		h.Send(chatID, "Введите сумму (числом), например: 1999.50", RecurringKeyboard())
		return
	}

	if draft["amount"].(float64) == 0.0 {
		// парсим сумму
		s := strings.ReplaceAll(text, ",", ".")
		amt, err := strconv.ParseFloat(s, 64)
		if err != nil || amt <= 0 {
			h.Send(chatID, "Некорректная сумма — введи число, например 1999.50", RecurringKeyboard())
			return
		}
		draft["amount"] = amt
		h.fsm.Set(userID, "RECURRING_ADD", draft)
		h.Send(chatID, "Введите период: daily / weekly / monthly / yearly", RecurringKeyboard())
		return
	}

	if draft["period"] == "" {
		p := strings.ToLower(text)
		if p != "daily" && p != "weekly" && p != "monthly" && p != "yearly" {
			h.Send(chatID, "Период должен быть одним из: daily, weekly, monthly, yearly", RecurringKeyboard())
			return
		}
		draft["period"] = p
		// если monthly — спросим день месяца (1-28 для безопасности)
		if p == "monthly" {
			h.fsm.Set(userID, "RECURRING_ADD", draft)
			h.Send(chatID, "Введите день месяца для платежа (1-28):", RecurringKeyboard())
			return
		}
		// иначе — сохраняем
		if err := h.saveRecurringFromDraft(userID, draft); err != nil {
			logger.Error("saveRecurring error: " + err.Error())
			h.Send(chatID, "Ошибка при сохранении регулярного платежа", RecurringKeyboard())
			h.fsm.Clear(userID)
			return
		}
		h.fsm.Clear(userID)
		h.Send(chatID, "Регулярный платёж сохранён.", RecurringKeyboard())
		return
	}

	// если мы сюда пришли — возможно period == monthly и нужно обработать payment_day
	if draft["period"] == "monthly" && draft["payment_day"].(int) == 0 {
		day, err := strconv.Atoi(text)
		if err != nil || day < 1 || day > 28 {
			h.Send(chatID, "Введите корректный день месяца: число от 1 до 28", RecurringKeyboard())
			return
		}
		draft["payment_day"] = day
		// сохраняем
		if err := h.saveRecurringFromDraft(userID, draft); err != nil {
			logger.Error("saveRecurring error: " + err.Error())
			h.Send(chatID, "Ошибка при сохранении регулярного платежа", RecurringKeyboard())
			h.fsm.Clear(userID)
			return
		}
		h.fsm.Clear(userID)
		h.Send(chatID, "Регулярный платёж сохранён.", RecurringKeyboard())
		return
	}

	// fallback
	h.Send(chatID, "Не удалось обработать сообщение. Попробуйте /start или Домой.", RecurringKeyboard())
}

// saveRecurringFromDraft — конвертация draft -> finance.RecurringPayment и сохранение в репо + создание todo-напоминания
func (h *Handler) saveRecurringFromDraft(userID int64, draft map[string]any) error {
	ctx := context.Background()

	title, _ := draft["title"].(string)
	amount, _ := draft["amount"].(float64)
	period, _ := draft["period"].(string)

	// вычислим next payment: простая логика — ближайшая дата в зависимости от периода
	now := time.Now().Truncate(24 * time.Hour)
	var next time.Time

	switch period {
	case "daily":
		next = now.AddDate(0, 0, 1)
	case "weekly":
		next = now.AddDate(0, 0, 7)
	case "monthly":
		day := 1
		if v, ok := draft["payment_day"].(int); ok && v >= 1 && v <= 28 {
			day = v
		}
		// если сегодня уже после day этого месяца — ставим на следующий месяц
		yy, mm, _ := now.Date()
		loc := now.Location()
		candidate := time.Date(yy, mm, day, 9, 0, 0, 0, loc)
		if !candidate.After(now) {
			// следующй месяц
			candidate = candidate.AddDate(0, 1, 0)
		}
		next = candidate
	case "yearly":
		next = now.AddDate(1, 0, 0)
	default:
		next = now.AddDate(0, 0, 1)
	}

	rp := &finance.RecurringPayment{
		UserID:      userID,
		Title:       title,
		Amount:      amount,
		Category:    "",
		Period:      period,
		NextPayment: next,
		CreatedAt:   time.Now(),
	}

	// Сохраняем через сервис финансов

	// Сохраняем через сервис
	_, err := h.finance.AddRecurring(ctx, rp)
	if err != nil {
		logger.Error("AddRecurring failed: " + err.Error())
		return err
	}

	// создание напоминания в todo — предупредить за 1 день (если next != today)
	remTitle := fmt.Sprintf("Платёж: %s — %.2f ₽", rp.Title, rp.Amount)
	// создаём напоминание за 1 день
	if err := h.todo.CreateAuto(ctx, userID, remTitle); err != nil {
		// логируем, но не ломаем основной процесс
		logger.Error("todo.CreateAuto failed: " + err.Error())
	}

	return nil
}

// showRecurringList — выводит список регулярных платежей (через recurringRepo)
func (h *Handler) showRecurringList(userID int64) {
	chatID := userID
	ctx := context.Background()

	// Получаем список через сервис
	list, err := h.finance.GetRecurringList(ctx, userID)
	if err != nil {
		logger.Error("GetRecurringList error: " + err.Error())
		h.Send(chatID, "Ошибка получения списка регулярных платежей", RecurringKeyboard())
		return
	}

	if len(list) == 0 {
		h.Send(chatID, "Регулярных платежей не найдено", RecurringKeyboard())
		return
	}

	var sb strings.Builder
	sb.WriteString("Регулярные платежи:\n\n")
	for _, p := range list {
		sb.WriteString(fmt.Sprintf("ID:%d • %s — %.2f ₽ • %s • next: %s\n",
			p.ID, p.Title, p.Amount, p.Period, p.NextPayment.Format("02.01.2006")))
	}
	sb.WriteString("\nДля удаления используйте /delete_recurring <id>")
	h.Send(chatID, sb.String(), RecurringKeyboard())
}

// handleDeleteRecurringCommand — /delete_recurring <id>
func (h *Handler) handleDeleteRecurringCommand(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		h.Send(chatID, "Использование: /delete_recurring <id>", RecurringKeyboard())
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		h.Send(chatID, "ID должен быть числом", RecurringKeyboard())
		return
	}

	// Удаляем через сервис
	if err := h.finance.DeleteRecurring(context.Background(), id); err != nil {
		logger.Error("DeleteRecurring failed: " + err.Error())
		h.Send(chatID, "Ошибка удаления регулярного платежа", RecurringKeyboard())
		return
	}

	h.Send(chatID, "Регулярный платёж удалён", RecurringKeyboard())
}
