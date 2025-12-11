package todo

import "time"

type Item struct {
	ID          int
	UserID      int64
	Title       string
	Description string
	DueDate     *time.Time
	Status      string
	CreatedAt   time.Time
}
