package credits

import "time"

// Credit представляет кредит пользователя.
type Credit struct {
	ID        int       // Уникальный идентификатор кредита
	UserID    int64     // ID пользователя-владельца
	Title     string    // Название кредита
	Principal float64   // Основная сумма кредита
	Rate      float64   // Годовая процентная ставка
	Months    int       // Срок кредита в месяцах
	CreatedAt time.Time // Дата создания
}

// Payment представляет один платёж по кредиту в графике платежей.
type Payment struct {
	DueDate   time.Time // Дата платежа
	Principal float64   // Часть основного долга
	Interest  float64   // Проценты
	Total     float64   // Общая сумма платежа
}
