package todo

import "context"

type Repository interface {
	Create(ctx context.Context, t *Item) (int, error)
	List(ctx context.Context, userID int64) ([]Item, error)
	Delete(ctx context.Context, userID int64, id int) error
}
