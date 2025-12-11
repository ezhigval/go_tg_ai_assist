package storage

import (
	"context"
	"time"

	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FinanceRepo struct {
	db *pgxpool.Pool
}

func NewFinanceRepo(db *pgxpool.Pool) *FinanceRepo {
	return &FinanceRepo{db: db}
}

func (r *FinanceRepo) AddEntry(ctx context.Context, e *finance.FinanceEntry) error {

	_, err := r.db.Exec(ctx, `
        INSERT INTO finance_entries (user_id, amount, category, type, note, created_at)
        VALUES ($1,$2,$3,$4,$5,$6)
    `,
		e.UserID, e.Amount, e.Category, e.Type, e.Note, time.Now(),
	)

	if err != nil {
		logger.Error("FinanceRepo.AddEntry error: " + err.Error())
	}

	return err
}

func (r *FinanceRepo) ListEntries(ctx context.Context, userID int64) ([]*finance.FinanceEntry, error) {

	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, amount, category, type, note, created_at
        FROM finance_entries
        WHERE user_id=$1
        ORDER BY created_at DESC
    `,
		userID,
	)

	if err != nil {
		logger.Error("FinanceRepo.ListEntries error: " + err.Error())
		return nil, err
	}

	defer rows.Close()

	var list []*finance.FinanceEntry

	for rows.Next() {
		var e finance.FinanceEntry

		if err := rows.Scan(&e.ID, &e.UserID, &e.Amount, &e.Category, &e.Type, &e.Note, &e.CreatedAt); err != nil {
			return nil, err
		}

		list = append(list, &e)
	}

	return list, nil
}
