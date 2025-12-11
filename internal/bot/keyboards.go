package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// HomeKeyboard возвращает главную клавиатуру бота с основными модулями.
func HomeKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdTodo),
			tgbotapi.NewKeyboardButton(CmdFinance),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdCredits),
		),
	)
}

// BackKeyboard возвращает клавиатуру с кнопками "Назад" и "Домой".
// Используется во время многошаговых диалогов (FSM).
func BackKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdBack),
			tgbotapi.NewKeyboardButton(CmdHome),
		),
	)
}

// TodoKeyboard возвращает клавиатуру модуля задач.
func TodoKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdTodoAdd),
			tgbotapi.NewKeyboardButton(CmdTodoList),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdBack),
			tgbotapi.NewKeyboardButton(CmdHome),
		),
	)
}

// FinanceKeyboard возвращает клавиатуру финансового модуля.
func FinanceKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdFinanceAdd),
			tgbotapi.NewKeyboardButton(CmdFinanceList),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdRecurring),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdBack),
			tgbotapi.NewKeyboardButton(CmdHome),
		),
	)
}

// RecurringKeyboard возвращает клавиатуру модуля регулярных платежей.
func RecurringKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdRecurringAdd),
			tgbotapi.NewKeyboardButton(CmdRecurringList),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdBack),
			tgbotapi.NewKeyboardButton(CmdHome),
		),
	)
}

// CreditsKeyboard возвращает клавиатуру кредитного модуля.
func CreditsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdCreditAdd),
			tgbotapi.NewKeyboardButton(CmdCreditList),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(CmdBack),
			tgbotapi.NewKeyboardButton(CmdHome),
		),
	)
}
