package storage

import (
	"context"
	"time"

	"tg_bot_asist/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) RegisterUser(ctx context.Context, userID, chatID int64, lastMessage string) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO users (user_id, chat_id, last_message, last_seen)
        VALUES ($1,$2,$3,$4)
        ON CONFLICT (user_id) DO UPDATE SET
            chat_id = EXCLUDED.chat_id,
            last_message = EXCLUDED.last_message,
            last_seen = EXCLUDED.last_seen
    `,
		userID, chatID, lastMessage, time.Now(),
	)

	if err != nil {
		logger.Error("UserRepo.RegisterUser error: " + err.Error())
	}

	return err
}
