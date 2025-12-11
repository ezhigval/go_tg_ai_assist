package storage

import (
	"context"
	"time"

	"tg_bot_asist/internal/credits"
	"tg_bot_asist/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CreditsRepo struct {
	db *pgxpool.Pool
}

func NewCreditsRepo(db *pgxpool.Pool) *CreditsRepo {
	return &CreditsRepo{db: db}
}

func (r *CreditsRepo) AddCredit(ctx context.Context, c *credits.Credit) (int, error) {

	var id int

	err := r.db.QueryRow(ctx, `
        INSERT INTO credits (user_id, title, principal, rate, months, created_at)
        VALUES ($1,$2,$3,$4,$5,$6)
        RETURNING id
    `,
		c.UserID, c.Title, c.Principal, c.Rate, c.Months, time.Now(),
	).Scan(&id)

	if err != nil {
		logger.Error("CreditsRepo.AddCredit error: " + err.Error())
		return 0, err
	}

	return id, nil
}

func (r *CreditsRepo) ListCredits(ctx context.Context, userID int64) ([]credits.Credit, error) {

	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, title, principal, rate, months, created_at
        FROM credits
        WHERE user_id=$1
        ORDER BY created_at DESC
    `,
		userID,
	)

	if err != nil {
		logger.Error("CreditsRepo.ListCredits error: " + err.Error())
		return nil, err
	}

	defer rows.Close()

	var list []credits.Credit

	for rows.Next() {
		var c credits.Credit
		err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Principal, &c.Rate, &c.Months, &c.CreatedAt)

		if err != nil {
			return nil, err
		}

		list = append(list, c)
	}

	return list, nil
}

// GetByID возвращает кредит по ID, если он принадлежит пользователю.
func (r *CreditsRepo) GetByID(ctx context.Context, id int, userID int64) (*credits.Credit, error) {
	var c credits.Credit
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, title, principal, rate, months, created_at
		FROM credits
		WHERE id=$1 AND user_id=$2
	`, id, userID).Scan(&c.ID, &c.UserID, &c.Title, &c.Principal, &c.Rate, &c.Months, &c.CreatedAt)

	if err != nil {
		logger.Error("CreditsRepo.GetByID error: " + err.Error())
		return nil, err
	}

	return &c, nil
}

// Delete удаляет кредит по ID, если он принадлежит пользователю.
func (r *CreditsRepo) Delete(ctx context.Context, id int, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM credits WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		logger.Error("CreditsRepo.Delete error: " + err.Error())
	}
	return err
}
