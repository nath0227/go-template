package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type gcsClient struct {
	client *storage.Client
	cfg    StorageConfig
}

func newGCSClient(cfg StorageConfig) (*gcsClient, error) {
	ctx := context.Background()

	var opts []option.ClientOption
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	zap.L().Info("gcs client created")
	return &gcsClient{client: client, cfg: cfg}, nil
}

func (c *gcsClient) Upload(ctx context.Context, bucket, key string, r io.Reader) error {
	wc := c.client.Bucket(bucket).Object(key).NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return err
	}
	return wc.Close()
}

func (c *gcsClient) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return c.client.Bucket(bucket).Object(key).NewReader(ctx)
}

func (c *gcsClient) Delete(ctx context.Context, bucket, key string) error {
	return c.client.Bucket(bucket).Object(key).Delete(ctx)
}

func (c *gcsClient) PublicURL(bucket, key string) string {
	if c.cfg.CDNBaseURL != "" {
		return fmt.Sprintf("%s/%s", c.cfg.CDNBaseURL, key)
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, key)
}

func (c *gcsClient) SignURL(_ context.Context, bucket, key string, ttl time.Duration) (string, error) {
	// Requires signing credentials: set STORAGE_CREDENTIALS_FILE to a service account key,
	// or run on GCE/Cloud Run where the metadata server provides an identity with signBlob permission.
	return c.client.Bucket(bucket).SignedURL(key, &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(ttl),
	})
}
