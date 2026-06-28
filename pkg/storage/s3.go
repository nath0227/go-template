package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type s3Client struct {
	client *s3.Client
	cfg    StorageConfig
}

func newS3Client(cfg StorageConfig) (*s3Client, error) {
	var opts []func(*awsconfig.LoadOptions) error
	opts = append(opts, awsconfig.WithRegion(cfg.Region))

	if cfg.AccessKeyID != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
	}

	zap.L().Info("s3 client created", zap.String("region", cfg.Region))
	return &s3Client{client: s3.NewFromConfig(awsCfg, s3Opts...), cfg: cfg}, nil
}

func (c *s3Client) Upload(ctx context.Context, bucket, key string, r io.Reader) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	return err
}

func (c *s3Client) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (c *s3Client) Delete(ctx context.Context, bucket, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (c *s3Client) PublicURL(bucket, key string) string {
	if c.cfg.CDNBaseURL != "" {
		return fmt.Sprintf("%s/%s", c.cfg.CDNBaseURL, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, c.cfg.Region, key)
}

func (c *s3Client) SignURL(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	pc := s3.NewPresignClient(c.client)
	req, err := pc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
