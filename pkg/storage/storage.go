package storage

import (
	"context"
	"fmt"
	"io"
)

type ObjectStorage interface {
	Upload(ctx context.Context, bucket, key string, r io.Reader) error
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	PublicURL(bucket, key string) string
}

type StorageConfig struct {
	Driver          string `env:"STORAGE_DRIVER" envDefault:"aws"`
	Region          string `env:"STORAGE_REGION"`
	Bucket          string `env:"STORAGE_BUCKET"`
	AccessKeyID     string `env:"STORAGE_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"STORAGE_SECRET_ACCESS_KEY"`
	// Endpoint overrides the default service endpoint (useful for MinIO or COS custom domains)
	Endpoint        string `env:"STORAGE_ENDPOINT"`
	// CredentialsFile is the path to a GCP service account JSON key file
	CredentialsFile string `env:"STORAGE_CREDENTIALS_FILE"`
	// CDNBaseURL replaces the default public URL with a CDN-fronted base
	CDNBaseURL      string `env:"STORAGE_CDN_BASE_URL"`
}

func NewStorage(cfg StorageConfig) (ObjectStorage, error) {
	switch cfg.Driver {
	case "aws":
		return newS3Client(cfg)
	case "gcp":
		return newGCSClient(cfg)
	case "tencent":
		return newCOSClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}
