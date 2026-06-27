package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/your-org/service-name/pkg/cache"
	pkgdb "github.com/your-org/service-name/pkg/db"
	"github.com/your-org/service-name/pkg/httpclient"
	"github.com/your-org/service-name/pkg/httpserver"
	"github.com/your-org/service-name/pkg/kafka"
	"github.com/your-org/service-name/pkg/logger"
	"github.com/your-org/service-name/pkg/storage"
)

type Config struct {
	DB      pkgdb.DBConfig
	Cache   cache.RedisConfig
	Kafka   kafka.KafkaConfig
	Storage storage.StorageConfig
	Client  httpclient.ClientConfig
	Server  httpserver.ServerConfig
	Log     logger.LogConfig
}

// Load reads configuration from config/.env (if present) then from environment variables.
// Defaults are embedded in each field's envDefault tag.
func Load() (*Config, error) {
	_ = godotenv.Load("config/.env")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
