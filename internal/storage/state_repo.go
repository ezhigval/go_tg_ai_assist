package storage

import (
	"context"

	"tg_bot_asist/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StateRepo struct {
	db *pgxpool.Pool
}

func NewStateRepo(db *pgxpool.Pool) *StateRepo {
	return &StateRepo{db: db}
}

func (r *StateRepo) Save(ctx context.Context, userID int64, state string) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO user_states (user_id, state_json)
        VALUES ($1,$2)
        ON CONFLICT (user_id) DO UPDATE SET state_json = EXCLUDED.state_json
    `,
		userID, state,
	)

	if err != nil {
		logger.Error("StateRepo.Save error: " + err.Error())
	}

	return err
}

func (r *StateRepo) Get(ctx context.Context, userID int64) (string, error) {
	var s string

	err := r.db.QueryRow(ctx, `
        SELECT state_json FROM user_states WHERE user_id=$1
    `, userID).Scan(&s)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", nil
		}
		logger.Error("StateRepo.Get error: " + err.Error())
		return "", err
	}

	return s, nil
}

func (r *StateRepo) Delete(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_states WHERE user_id=$1`, userID)

	if err != nil {
		logger.Error("StateRepo.Delete error: " + err.Error())
	}

	return err
}
