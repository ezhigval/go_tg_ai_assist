package bot

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"tg_bot_asist/internal/credits"
	"tg_bot_asist/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// startCreditAdd — запускает FSM для добавления нового кредита
func (h *Handler) startCreditAdd(userID int64) {
	h.fsm.Set(userID, "CREDIT_ADD", map[string]any{})
	h.Send(userID, "Введите название кредита:", BackKeyboard())
}

// handleCreditAdd — FSM: собирает данные по шагам и сохраняет кредит через сервис
func (h *Handler) handleCreditAdd(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	state := h.fsm.Get(userID)
	if state == nil || state.Name != "CREDIT_ADD" {
		h.Send(chatID, "Нет активного процесса добавления кредита", CreditsKeyboard())
		return
	}

	// state.Data уже map[string]any, используем напрямую
	data := state.Data
	if data == nil {
		data = map[string]any{}
	}

	// шаги: title -> principal -> rate -> months -> confirm
	step, _ := data["step"].(string)
	switch step {
	case "", "title":
		// записываем название
		data["title"] = text
		data["step"] = "principal"
		h.fsm.Set(userID, "CREDIT_ADD", data)
		h.Send(chatID, "Введите сумму кредита (числом):", BackKeyboard())
		return

	case "principal":
		p, err := strconv.ParseFloat(strings.ReplaceAll(text, ",", "."), 64)
		if err != nil || p <= 0 {
			h.Send(chatID, "Неверная сумма. Введите сумму (например 1230000):", BackKeyboard())
			return
		}
		data["principal"] = p
		data["step"] = "rate"
		h.fsm.Set(userID, "CREDIT_ADD", data)
		h.Send(chatID, "Введите годовую процентную ставку (например 12.5):", BackKeyboard())
		return

	case "rate":
		r, err := strconv.ParseFloat(strings.ReplaceAll(text, ",", "."), 64)
		if err != nil || r < 0 {
			h.Send(chatID, "Неверная ставка. Введите число (например 12.5):", BackKeyboard())
			return
		}
		data["rate"] = r
		data["step"] = "months"
		h.fsm.Set(userID, "CREDIT_ADD", data)
		h.Send(chatID, "Введите срок кредита в месяцах (например 36):", BackKeyboard())
		return

	case "months":
		m, err := strconv.Atoi(text)
		if err != nil || m <= 0 {
			h.Send(chatID, "Неверный срок. Введите целое число месяцев:", BackKeyboard())
			return
		}
		data["months"] = m
		// финальный шаг — сохраняем
		credit := &credits.Credit{
			UserID:    userID,
			Title:     data["title"].(string),
			Principal: data["principal"].(float64),
			Rate:      data["rate"].(float64),
			Months:    m,
		}

		id, err := h.credits.Add(context.Background(), credit)
		if err != nil {
			logger.Error("credit save error: " + err.Error())
			h.Send(chatID, "Ошибка при сохранении кредита", CreditsKeyboard())
			h.fsm.Clear(userID)
			return
		}

		msg := fmt.Sprintf("Кредит сохранён. ID: %d\n%s — %.2f ₽, ставка %.2f%%, %d мес.",
			id, credit.Title, credit.Principal, credit.Rate, credit.Months)
		h.Send(chatID, msg, CreditsKeyboard())
		h.fsm.Clear(userID)
		return
	}
}

// showCreditList — выводит список кредитов пользователя
func (h *Handler) showCreditList(userID int64) {
	chatID := userID
	ctx := context.Background()

	list, err := h.credits.List(ctx, userID)
	if err != nil {
		logger.Error("credits list error: " + err.Error())
		h.Send(chatID, "Ошибка получения списка кредитов", CreditsKeyboard())
		return
	}

	if len(list) == 0 {
		h.Send(chatID, "Кредитов нет", CreditsKeyboard())
		return
	}

	var b strings.Builder
	b.WriteString("Ваши кредиты:\n\n")
	for _, c := range list {
		line := fmt.Sprintf("ID:%d • %s\nСумма: %.2f ₽ • %.2f%% • %d мес.\n\n", c.ID, c.Title, c.Principal, c.Rate, c.Months)
		b.WriteString(line)
	}
	b.WriteString("Чтобы посмотреть график платежей: /payments <id>\nЧтобы скопировать кредит: /copy_credit <id>\nЧтобы закрыть кредит: /close_credit <id>")
	h.Send(chatID, b.String(), CreditsKeyboard())
}

// handlePaymentsCommand — парсит /payments <id> и отправляет график
func (h *Handler) handlePaymentsCommand(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		h.Send(chatID, "Использование: /payments <id>", CreditsKeyboard())
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		h.Send(chatID, "ID должен быть числом", CreditsKeyboard())
		return
	}

	// Получаем кредит по ID через сервис
	ctx := context.Background()
	found, err := h.credits.GetByID(ctx, id, update.Message.From.ID)
	if err != nil {
		logger.Error("credits GetByID error: " + err.Error())
		h.Send(chatID, "Кредит не найден", CreditsKeyboard())
		return
	}

	// считаем график (аннуитет)
	schedule := calcAnnuitySchedule(found.Principal, found.Rate, found.Months, time.Now())

	// форматируем и отправляем первые N строк (чтобы не спамить)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("График платежей по кредиту %s (ID:%d)\n\n", found.Title, found.ID))
	total := 0.0
	for i, p := range schedule {
		b.WriteString(fmt.Sprintf("%2d) %s — %.2f (обр: %.2f, %s: %.2f)\n",
			i+1,
			p.Date.Format("02.01.2006"),
			p.Payment,
			p.PrincipalPart,
			map[string]string{"interest": "Проценты", "principal": "Основной"}[p.Kind],
			p.InterestPart,
		))
		total += p.Payment
		if i >= 11 { // показываем первые 12 платежей, по желанию можно показать весь график файлом
			b.WriteString(fmt.Sprintf("\nИтого (первые 12): %.2f ₽\n", total))
			break
		}
	}

	h.Send(chatID, b.String(), CreditsKeyboard())
}

// calcAnnuitySchedule — возвращает список платежей (аннуитет)
// paymentDate baseDate + i months
type PaymentRow struct {
	Date          time.Time
	Payment       float64
	PrincipalPart float64
	InterestPart  float64
	Remaining     float64
	Kind          string // "interest" or "principal" (aux)
}

func calcAnnuitySchedule(principal float64, annualRate float64, months int, baseDate time.Time) []PaymentRow {
	if months <= 0 {
		return nil
	}
	monthlyRate := annualRate / 100.0 / 12.0
	var payment float64
	if monthlyRate == 0 {
		payment = principal / float64(months)
	} else {
		r := monthlyRate
		payment = principal * (r * mathPow(1+r, months)) / (mathPow(1+r, months) - 1)
	}

	remaining := principal
	out := make([]PaymentRow, 0, months)
	for i := 0; i < months; i++ {
		interest := remaining * monthlyRate
		principalPart := payment - interest
		if principalPart > remaining {
			principalPart = remaining
		}
		remaining -= principalPart
		row := PaymentRow{
			Date:          baseDate.AddDate(0, i+1, 0),
			Payment:       round(payment, 2),
			PrincipalPart: round(principalPart, 2),
			InterestPart:  round(interest, 2),
			Remaining:     round(remaining, 2),
			Kind:          "payment",
		}
		out = append(out, row)
		if remaining <= 0 {
			break
		}
	}
	return out
}

// helper: fast pow
func mathPow(a float64, b int) float64 {
	return mathFloatPow(a, float64(b))
}

// use standard lib under alias to avoid naming conflicts
func mathFloatPow(x, y float64) float64 {
	return math.Pow(x, y)
}

func round(x float64, prec int) float64 {
	p := math.Pow(10, float64(prec))
	return math.Round(x*p) / p
}

// copyCredit — создаёт копию кредита (useful to duplicate schemes)
func (h *Handler) copyCreditCommand(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		h.Send(chatID, "Использование: /copy_credit <id>", CreditsKeyboard())
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		h.Send(chatID, "ID должен быть числом", CreditsKeyboard())
		return
	}

	ctx := context.Background()
	newID, err := h.credits.Copy(ctx, id, update.Message.From.ID)
	if err != nil {
		logger.Error("credit copy error: " + err.Error())
		h.Send(chatID, "Ошибка копирования кредита: "+err.Error(), CreditsKeyboard())
		return
	}

	h.Send(chatID, fmt.Sprintf("Кредит скопирован, новый ID: %d", newID), CreditsKeyboard())
}

// closeCreditCommand — удаляет кредит (простая операция)
func (h *Handler) closeCreditCommand(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		h.Send(chatID, "Использование: /close_credit <id>", CreditsKeyboard())
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		h.Send(chatID, "ID должен быть числом", CreditsKeyboard())
		return
	}

	// Удаляем кредит через сервис
	if err := h.credits.Delete(context.Background(), id, update.Message.From.ID); err != nil {
		logger.Error("credit close error: " + err.Error())
		h.Send(chatID, "Ошибка закрытия кредита", CreditsKeyboard())
		return
	}

	h.Send(chatID, "Кредит закрыт (удалён).", CreditsKeyboard())
}
