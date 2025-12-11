package bot

import (
	"strings"

	"tg_bot_asist/internal/credits"
	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/todo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbotapi.BotAPI
	fsm     *FSM
	todo    *todo.Service
	finance *finance.Service
	credits *credits.Service
}

func NewHandler(
	b *tgbotapi.BotAPI,
	fsm *FSM,
	todoSvc *todo.Service,
	finSvc *finance.Service,
	credSvc *credits.Service,
) *Handler {
	return &Handler{
		bot:     b,
		fsm:     fsm,
		todo:    todoSvc,
		finance: finSvc,
		credits: credSvc,
	}
}

// Send отправляет текстовое сообщение с клавиатурой пользователю.
func (h *Handler) Send(chatID int64, text string, kb tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	if _, err := h.bot.Send(msg); err != nil {
		logger.Error("Failed to send message: " + err.Error())
	}
}

// Edit отправляет сообщение с клавиатурой (используется вместо редактирования для ReplyKeyboard).
func (h *Handler) Edit(chatID int64, messageID int, text string, kb tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	if _, err := h.bot.Send(msg); err != nil {
		logger.Error("Failed to send message: " + err.Error())
	}
}

func (h *Handler) Handle(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	text := update.Message.Text

	// Глобальные кнопки
	if text == CmdHome {
		h.fsm.Clear(userID)
		h.Send(userID, "Главное меню", HomeKeyboard())
		return
	}

	if text == CmdBack {
		h.fsm.Clear(userID)
		h.Send(userID, "Назад", HomeKeyboard())
		return
	}

	// FSM → переслать в нужный модуль
	state := h.fsm.Get(userID)
	if state != nil {
		switch state.Name {
		case "TODO_ADD":
			h.handleTodoAdd(update)
			return
		case "FIN_ADD_TYPE", "FIN_ADD_AMOUNT", "FIN_ADD_CATEGORY", "FIN_ADD_DESC":
			h.handleFinanceFSM(update)
			return
		case "RECURRING_ADD":
			h.handleRecurringAdd(update)
			return
		case "CREDIT_ADD":
			h.handleCreditAdd(update)
			return
		}
	}

	// Команды первого уровня
	switch text {

	case CmdStart:
		h.Send(userID, "Привет! Я бот-помощник.\n\nИспользуйте меню для навигации или откройте веб-интерфейс через кнопку меню.", HomeKeyboard())

	case CmdTodo:
		h.Send(userID, "Модуль задач", TodoKeyboard())

	case CmdFinance:
		h.Send(userID, "Финансовый модуль", FinanceKeyboard())

	case CmdCredits:
		h.Send(userID, "Кредитный модуль", CreditsKeyboard())

	// TODO
	case CmdTodoAdd:
		h.startTodoAdd(userID)
	case CmdTodoList:
		h.showTodoList(userID)

	// FINANCE
	case CmdFinanceAdd:
		h.financeStart(userID, 0)
	case CmdFinanceList:
		h.showFinanceList(userID)

	case CmdRecurring:
		h.Send(userID, "Регулярные платежи", RecurringKeyboard())
	case CmdRecurringAdd:
		h.startRecurringAdd(userID)
	case CmdRecurringList:
		h.showRecurringList(userID)

	// CREDITS
	case CmdCreditAdd:
		h.startCreditAdd(userID)
	case CmdCreditList:
		h.showCreditList(userID)
	default:
		// Обработка команд с префиксом /
		if strings.HasPrefix(text, "/payments") {
			h.handlePaymentsCommand(update)
		} else if strings.HasPrefix(text, "/copy_credit") {
			h.copyCreditCommand(update)
		} else if strings.HasPrefix(text, "/close_credit") {
			h.closeCreditCommand(update)
		} else if strings.HasPrefix(text, "/delete_recurring") {
			h.handleDeleteRecurringCommand(update)
		}
	}
}

// Методы для TODO реализованы в todo_impl.go
// Методы для Recurring реализованы в recurring_impl.go
// Методы для Credits реализованы в credit_impl.go
