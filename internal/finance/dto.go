package finance

import "time"

type FinanceEntry struct {
	ID        int
	UserID    int64
	Amount    float64
	Category  string
	Type      string
	Note      string
	CreatedAt time.Time
}

type RecurringPayment struct {
	ID          int
	UserID      int64
	Title       string
	Amount      float64
	Category    string
	Period      string
	NextPayment time.Time
	CreatedAt   time.Time
}
