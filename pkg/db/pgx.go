package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewPGX(cfg DBConfig) (*pgxpool.Pool, error) {
	dsn := cfg.dsn()
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgx ping failed: %w", err)
	}
	zap.L().Info("pgx pool connected",
		zap.String("host", cfg.Host),
		zap.String("db", cfg.Name),
	)
	return pool, nil
}
