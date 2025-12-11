package config

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"tg_bot_asist/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() (*pgxpool.Pool, error) {
	host := Get("DB_HOST")
	port := Get("DB_PORT")
	user := Get("DB_USER")
	pass := Get("DB_PASSWORD")
	name := Get("DB_NAME")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user,
		pass,
		host,
		port,
		name,
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if poolSize := Get("DB_POOL_SIZE"); poolSize != "" {
		if n, err := strconv.Atoi(poolSize); err == nil {
			cfg.MaxConns = int32(n)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	logger.Info("Database connection established")
	return db, nil
}
