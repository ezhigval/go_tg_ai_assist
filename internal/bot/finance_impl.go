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

// Вспомогательные функции для работы с map[string]any
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]any, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// =============================
// FSM: ADD_FINANCE
// =============================

type financeDraft struct {
	Type        string // income / expense
	Category    string
	Amount      float64
	Description string
	Date        time.Time
}

// financeStart создаёт первую ступень FSM: выбор доход/расход.
func (h *Handler) financeStart(userID int64, msgID int) {
	draft := &financeDraft{}
	h.fsm.Set(userID, "FIN_ADD_TYPE", map[string]any{
		"type":        draft.Type,
		"category":    draft.Category,
		"amount":      draft.Amount,
		"description": draft.Description,
	})
	h.Send(userID, "Выберите тип операции:\nДоход или Расход", BackKeyboard())
}

// Обработка всех этапов FSM по фин. операциям
func (h *Handler) handleFinanceFSM(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	text := strings.TrimSpace(update.Message.Text)
	state := h.fsm.Get(userID)
	if state == nil {
		return
	}

	// state.Data уже map[string]any, используем напрямую
	data := state.Data
	if data == nil {
		return
	}

	// Восстанавливаем draft из map
	draft := &financeDraft{
		Type:        getString(data, "type"),
		Category:    getString(data, "category"),
		Amount:      getFloat(data, "amount"),
		Description: getString(data, "description"),
	}

	switch state.Name {

	case "FIN_ADD_TYPE":
		// user picked income/expense
		if text == "Доход" {
			draft.Type = "income"
		} else if text == "Расход" {
			draft.Type = "expense"
		} else {
			h.Send(userID, "Выберите тип: Доход или Расход", BackKeyboard())
			return
		}

		h.fsm.Set(userID, "FIN_ADD_AMOUNT", map[string]any{
			"type":        draft.Type,
			"category":    draft.Category,
			"amount":      draft.Amount,
			"description": draft.Description,
		})
		h.Send(userID, "Введите сумму операции:", BackKeyboard())
		return

	case "FIN_ADD_AMOUNT":
		val, err := strconv.ParseFloat(text, 64)
		if err != nil || val <= 0 {
			h.Send(userID, "Введите корректную сумму:", BackKeyboard())
			return
		}
		draft.Amount = val

		h.fsm.Set(userID, "FIN_ADD_CATEGORY", map[string]any{
			"type":        draft.Type,
			"category":    draft.Category,
			"amount":      draft.Amount,
			"description": draft.Description,
		})
		h.Send(userID, "Введите категорию операции:", BackKeyboard())
		return

	case "FIN_ADD_CATEGORY":
		if text == "" {
			h.Send(userID, "Категория не может быть пустой. Введите категорию:", BackKeyboard())
			return
		}
		draft.Category = text

		h.fsm.Set(userID, "FIN_ADD_DESC", map[string]any{
			"type":        draft.Type,
			"category":    draft.Category,
			"amount":      draft.Amount,
			"description": draft.Description,
		})
		h.Send(userID, "Введите описание (или '-' если нет):", BackKeyboard())
		return

	case "FIN_ADD_DESC":
		if text != "-" {
			draft.Description = text
		}

		// Готовим финальный шаг: сохраняем
		h.saveFinance(userID, draft)
		h.fsm.Clear(userID)
		return
	}
}

func (h *Handler) saveFinance(userID int64, d *financeDraft) {
	ctx := context.Background()

	entry := &finance.FinanceEntry{
		UserID:    userID,
		Amount:    d.Amount,
		Category:  d.Category,
		Type:      d.Type,
		Note:      d.Description,
		CreatedAt: time.Now(),
	}

	err := h.finance.AddEntry(ctx, entry)
	if err != nil {
		logger.Error("Finance save error: " + err.Error())
		h.Send(userID, "Ошибка при сохранении операции", FinanceKeyboard())
		return
	}

	msg := fmt.Sprintf(
		"%s на сумму %.2f ₽ сохранён.\nКатегория: %s",
		map[string]string{"income": "Доход", "expense": "Расход"}[d.Type],
		d.Amount,
		d.Category,
	)

	h.Send(userID, msg, FinanceKeyboard())
}

// ===========================
// Список операций
// ===========================

func (h *Handler) showFinanceList(userID int64) {
	ctx := context.Background()

	ops, err := h.finance.ListEntries(ctx, userID)
	if err != nil {
		logger.Error("Finance list error: " + err.Error())
		h.Send(userID, "Ошибка получения списка финансов", FinanceKeyboard())
		return
	}

	if len(ops) == 0 {
		h.Send(userID, "У вас пока нет операций.", FinanceKeyboard())
		return
	}

	var b strings.Builder
	b.WriteString("Ваши операции:\n\n")

	totalIncome := 0.0
	totalExpense := 0.0

	for _, op := range ops {
		sign := "+"
		if op.Type == "expense" {
			sign = "-"
			totalExpense += op.Amount
		} else {
			totalIncome += op.Amount
		}

		line := fmt.Sprintf(
			"%s %.2f ₽ | %s | %s\n",
			sign,
			op.Amount,
			op.Category,
			op.CreatedAt.Format("02.01.2006"),
		)
		b.WriteString(line)
	}

	b.WriteString("\nИтоги:\n")
	b.WriteString(fmt.Sprintf("Доходы: %.2f ₽\n", totalIncome))
	b.WriteString(fmt.Sprintf("Расходы: %.2f ₽\n", totalExpense))
	b.WriteString(fmt.Sprintf("Баланс: %.2f ₽\n", totalIncome-totalExpense))

	h.Send(userID, b.String(), FinanceKeyboard())
}
