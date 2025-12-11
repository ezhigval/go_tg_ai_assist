package credits

import "context"

// Repository определяет интерфейс для работы с кредитами в хранилище.
type Repository interface {
	AddCredit(ctx context.Context, c *Credit) (int, error)
	ListCredits(ctx context.Context, userID int64) ([]Credit, error)
	GetByID(ctx context.Context, id int, userID int64) (*Credit, error)
	Delete(ctx context.Context, id int, userID int64) error
}
