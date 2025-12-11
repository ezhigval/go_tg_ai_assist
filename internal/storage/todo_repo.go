package storage

import (
	"context"
	"time"

	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/todo"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepo struct {
	db *pgxpool.Pool
}

func NewTodoRepo(db *pgxpool.Pool) *TodoRepo {
	return &TodoRepo{db: db}
}

func (r *TodoRepo) Create(ctx context.Context, t *todo.Item) (int, error) {

	var id int

	err := r.db.QueryRow(ctx, `
        INSERT INTO todos (user_id, title, description, due_date, status, created_at)
        VALUES ($1,$2,$3,$4,$5,$6)
        RETURNING id
    `,
		t.UserID, t.Title, t.Description, t.DueDate, t.Status, time.Now(),
	).Scan(&id)

	if err != nil {
		logger.Error("TodoRepo.Create error: " + err.Error())
		return 0, err
	}

	return id, nil
}

func (r *TodoRepo) List(ctx context.Context, userID int64) ([]todo.Item, error) {

	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, title, description, due_date, status, created_at
        FROM todos
        WHERE user_id=$1
        ORDER BY created_at DESC
    `,
		userID,
	)

	if err != nil {
		logger.Error("TodoRepo.List error: " + err.Error())
		return nil, err
	}

	defer rows.Close()

	var list []todo.Item

	for rows.Next() {
		var it todo.Item
		err := rows.Scan(&it.ID, &it.UserID, &it.Title, &it.Description, &it.DueDate, &it.Status, &it.CreatedAt)

		if err != nil {
			return nil, err
		}

		list = append(list, it)
	}

	return list, nil
}

func (r *TodoRepo) Delete(ctx context.Context, userID int64, id int) error {

	_, err := r.db.Exec(ctx,
		`DELETE FROM todos WHERE id=$1 AND user_id=$2`,
		id, userID,
	)

	if err != nil {
		logger.Error("TodoRepo.Delete error: " + err.Error())
	}

	return err
}
