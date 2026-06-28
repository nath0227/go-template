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

// DBConnConfig holds connection credentials for a single DB endpoint (writer or reader).
type DBConnConfig struct {
	Host     string `env:"HOST"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	Name     string `env:"NAME"`
	SSLMode  string `env:"SSL_MODE" envDefault:"disable"`
	// MySQL only
	Timezone string `env:"TIMEZONE" envDefault:"UTC"`
	// PostgreSQL only: appears in pg_stat_activity; useful for identifying writer vs reader
	ApplicationName string `env:"APPLICATION_NAME"`
	// PostgreSQL only: sets search_path
	Schema string `env:"SCHEMA"`
}

// postgresDSN builds a libpq-compatible connection string.
func (c DBConnConfig) postgresDSN() string {
	params := url.Values{}
	params.Set("sslmode", c.SSLMode)
	if c.ApplicationName != "" {
		params.Set("application_name", c.ApplicationName)
	}
	if c.Schema != "" {
		params.Set("search_path", c.Schema)
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s?%s",
		c.User, c.Password, c.Host, c.Name, params.Encode(),
	)
}

// BackendType selects which DB client to initialize.
type BackendType string

const (
	BackendGORM BackendType = "gorm"
	BackendPGX  BackendType = "pgx"
	BackendBoth BackendType = "both"
)

// DBConfig holds the full database configuration: writer, optional reader, shared pool settings.
type DBConfig struct {
	Backend BackendType `env:"DB_BACKEND" envDefault:"gorm"`
	Driver  string      `env:"DB_DRIVER" envDefault:"postgres"`
	// Writer is the primary (read-write) connection.
	Writer DBConnConfig `envPrefix:"DB_WRITER_"`
	// Reader is the replica (read-only) connection.
	// When DB_READER_HOST is empty, reads fall back to Writer.
	Reader DBConnConfig `envPrefix:"DB_READER_"`
	// Pool settings are shared between writer and reader connections.
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS"    envDefault:"10"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS"    envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
}

func (c DBConfig) dsn(conn DBConnConfig) string {
	if c.Driver == "mysql" {
		return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True&loc=%s",
			conn.User, conn.Password, conn.Host, conn.Name, conn.Timezone,
		)
	}
	return conn.postgresDSN()
}

// DB holds separate GORM instances for the writer (primary) and reader (replica).
type DB struct {
	Writer *gorm.DB
	Reader *gorm.DB
}

// NewGORM opens the writer connection and, when DB_READER_HOST is set, the reader connection.
// Reader falls back to Writer when no reader is configured.
func NewGORM(cfg DBConfig) (*DB, error) {
	writer, err := openGORM(cfg, cfg.Writer)
	if err != nil {
		return nil, err
	}

	reader := writer
	if cfg.Reader.Host != "" {
		reader, err = openGORM(cfg, cfg.Reader)
		if err != nil {
			return nil, err
		}
	}

	return &DB{Writer: writer, Reader: reader}, nil
}

func openGORM(cfg DBConfig, conn DBConnConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.dsn(conn))
	default:
		dialector = postgres.Open(cfg.dsn(conn))
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
		return nil, fmt.Errorf("db ping failed (%s): %w", conn.Host, err)
	}

	zap.L().Info("gorm connected",
		zap.String("driver", cfg.Driver),
		zap.String("host", conn.Host),
		zap.String("db", conn.Name),
	)
	return db, nil
}
