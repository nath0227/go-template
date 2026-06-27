package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	cos "github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"
)

type cosClient struct {
	cfg StorageConfig
}

func newCOSClient(cfg StorageConfig) (*cosClient, error) {
	zap.L().Info("cos client created", zap.String("region", cfg.Region))
	return &cosClient{cfg: cfg}, nil
}

func (c *cosClient) clientFor(bucket string) *cos.Client {
	bucketURL, _ := url.Parse(
		fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, c.cfg.Region),
	)
	return cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.cfg.AccessKeyID,
			SecretKey: c.cfg.SecretAccessKey,
		},
	})
}

func (c *cosClient) Upload(ctx context.Context, bucket, key string, r io.Reader) error {
	_, err := c.clientFor(bucket).Object.Put(ctx, key, r, nil)
	return err
}

func (c *cosClient) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	resp, err := c.clientFor(bucket).Object.Get(ctx, key, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *cosClient) Delete(ctx context.Context, bucket, key string) error {
	_, err := c.clientFor(bucket).Object.Delete(ctx, key)
	return err
}

func (c *cosClient) PublicURL(bucket, key string) string {
	if c.cfg.CDNBaseURL != "" {
		return fmt.Sprintf("%s/%s", c.cfg.CDNBaseURL, key)
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", bucket, c.cfg.Region, key)
}
