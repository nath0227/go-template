package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PGXPool holds separate pgx pools for the writer (primary) and reader (replica).
type PGXPool struct {
	Writer *pgxpool.Pool
	Reader *pgxpool.Pool
}

// NewPGX opens the writer pool and, when DB_READER_HOST is set, the reader pool.
// Reader falls back to Writer when no reader is configured.
func NewPGX(cfg DBConfig) (*PGXPool, error) {
	writer, err := newPGXPool(cfg.Writer)
	if err != nil {
		return nil, err
	}

	reader := writer
	if cfg.Reader.Host != "" {
		reader, err = newPGXPool(cfg.Reader)
		if err != nil {
			return nil, err
		}
	}

	return &PGXPool{Writer: writer, Reader: reader}, nil
}

func newPGXPool(conn DBConnConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), conn.postgresDSN())
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgx ping failed (%s): %w", conn.Host, err)
	}
	zap.L().Info("pgx pool connected",
		zap.String("host", conn.Host),
		zap.String("db", conn.Name),
	)
	return pool, nil
}
