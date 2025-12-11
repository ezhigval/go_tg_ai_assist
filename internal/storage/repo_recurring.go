package storage

import (
	"context"
	"time"

	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecurringRepo struct {
	db *pgxpool.Pool
}

func NewRecurringRepo(db *pgxpool.Pool) *RecurringRepo {
	return &RecurringRepo{db: db}
}

func (r *RecurringRepo) Add(ctx context.Context, p *finance.RecurringPayment) (int, error) {

	var id int

	err := r.db.QueryRow(ctx, `
        INSERT INTO recurring_payments (user_id, title, amount, category, period, next_payment, created_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        RETURNING id
    `,
		p.UserID, p.Title, p.Amount, p.Category, p.Period, p.NextPayment, time.Now(),
	).Scan(&id)

	if err != nil {
		logger.Error("RecurringRepo.Add error: " + err.Error())
		return 0, err
	}

	return id, nil
}

func (r *RecurringRepo) List(ctx context.Context, userID int64) ([]finance.RecurringPayment, error) {

	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, title, amount, category, period, next_payment, created_at
        FROM recurring_payments
        WHERE user_id=$1
        ORDER BY next_payment
    `,
		userID,
	)

	if err != nil {
		logger.Error("RecurringRepo.List error: " + err.Error())
		return nil, err
	}

	defer rows.Close()

	var list []finance.RecurringPayment

	for rows.Next() {
		var p finance.RecurringPayment

		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Amount, &p.Category, &p.Period, &p.NextPayment, &p.CreatedAt); err != nil {
			return nil, err
		}

		list = append(list, p)
	}

	return list, nil
}

func (r *RecurringRepo) Delete(ctx context.Context, id int) error {

	_, err := r.db.Exec(ctx, `DELETE FROM recurring_payments WHERE id=$1`, id)

	if err != nil {
		logger.Error("RecurringRepo.Delete error: " + err.Error())
	}

	return err
}

func (r *RecurringRepo) UpdateNextPayment(ctx context.Context, id int, next time.Time) error {

	_, err := r.db.Exec(ctx,
		`UPDATE recurring_payments SET next_payment=$2 WHERE id=$1`,
		id, next,
	)

	if err != nil {
		logger.Error("RecurringRepo.UpdateNextPayment error: " + err.Error())
	}

	return err
}
