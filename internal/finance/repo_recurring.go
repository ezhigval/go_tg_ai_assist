package finance

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecurringRepo struct {
	db *pgxpool.Pool
}

func NewRecurringRepo(db *pgxpool.Pool) *RecurringRepo {
	return &RecurringRepo{db: db}
}

func (r *RecurringRepo) Add(ctx context.Context, p *RecurringPayment) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `
        INSERT INTO recurring_payments (user_id, title, amount, category, period, next_payment, created_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        RETURNING id
    `,
		p.UserID, p.Title, p.Amount, p.Category, p.Period, p.NextPayment, time.Now(),
	).Scan(&id)
	return id, err
}

func (r *RecurringRepo) List(ctx context.Context, userID int64) ([]RecurringPayment, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, title, amount, category, period, next_payment, created_at
        FROM recurring_payments
        WHERE user_id=$1
        ORDER BY next_payment
    `,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []RecurringPayment
	for rows.Next() {
		var p RecurringPayment
		rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Amount, &p.Category, &p.Period, &p.NextPayment, &p.CreatedAt)
		list = append(list, p)
	}
	return list, nil
}

func (r *RecurringRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM recurring_payments WHERE id=$1`, id)
	return err
}

func (r *RecurringRepo) UpdateNextPayment(ctx context.Context, id int, next time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE recurring_payments SET next_payment=$2 WHERE id=$1`, id, next)
	return err
}

// GetUserPayments возвращает все регулярные платежи пользователя.
// Если userID == 0, возвращает платежи всех пользователей (для scheduler).
func (r *RecurringRepo) GetUserPayments(ctx context.Context, userID int64) ([]*RecurringPayment, error) {
	var rows pgx.Rows
	var err error

	if userID == 0 {
		// Для scheduler: получаем все платежи
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, title, amount, category, period, next_payment, created_at
			FROM recurring_payments
			ORDER BY next_payment
		`)
	} else {
		// Для конкретного пользователя
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, title, amount, category, period, next_payment, created_at
			FROM recurring_payments
			WHERE user_id=$1
			ORDER BY next_payment
		`, userID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*RecurringPayment

	for rows.Next() {
		var p RecurringPayment
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Amount, &p.Category, &p.Period, &p.NextPayment, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}

	return list, nil
}
