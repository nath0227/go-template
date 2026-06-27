package db

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	Driver string `env:"DB_DRIVER" envDefault:"postgres"`
	// Host must include the port, e.g. "localhost:5432" or "localhost:3306"
	Host     string `env:"DB_HOST" envDefault:"localhost:5432"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
	// Timezone is MySQL-only; ignored for postgres
	Timezone string `env:"DB_TIMEZONE" envDefault:"UTC"`
	// ApplicationName appears in pg_stat_activity and slow-query logs; useful for identifying the service.
	ApplicationName string `env:"DB_APPLICATION_NAME"`
	// Schema sets the PostgreSQL search_path (e.g. "myschema"). Ignored for MySQL.
	Schema          string        `env:"DB_SCHEMA"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
}

func (c DBConfig) dsn() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?parseTime=True&loc=%s",
			c.User, c.Password, c.Host, c.Name, c.Timezone,
		)
	default:
		params := url.Values{}
		params.Set("sslmode", c.SSLMode)
		if c.ApplicationName != "" {
			params.Set("application_name", c.ApplicationName)
		}
		if c.Schema != "" {
			params.Set("search_path", c.Schema)
		}
		return fmt.Sprintf(
			"postgres://%s:%s@%s/%s?%s",
			c.User, c.Password, c.Host, c.Name, params.Encode(),
		)
	}
}

func NewGORM(cfg DBConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.dsn())
	default:
		dialector = postgres.Open(cfg.dsn())
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	zap.L().Info("gorm connected",
		zap.String("driver", cfg.Driver),
		zap.String("host", cfg.Host),
		zap.String("db", cfg.Name),
	)
	return db, nil
}
