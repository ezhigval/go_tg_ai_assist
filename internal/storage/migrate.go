package storage

import (
	"context"
	"embed"

	"tg_bot_asist/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {

	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, f := range files {
		name := f.Name()

		content, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}

		logger.Info("Running migration: " + name)

		if _, err := db.Exec(ctx, string(content)); err != nil {
			logger.Error("Migration failed: " + err.Error())
			return err
		}
	}

	return nil
}
