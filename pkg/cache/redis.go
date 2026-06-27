package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
	PoolSize int    `env:"REDIS_POOL_SIZE" envDefault:"10"`
}

func NewRedis(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	zap.L().Info("redis connected", zap.String("addr", cfg.Addr))
	return client, nil
}
