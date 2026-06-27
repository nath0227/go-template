package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/your-org/service-name/pkg/logger"
)

type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"
	ContentTypeForm ContentType = "application/x-www-form-urlencoded"
)

type ClientConfig struct {
	Timeout         time.Duration `env:"HTTP_CLIENT_TIMEOUT" envDefault:"30s"`
	MaxIdleConns    int           `env:"HTTP_CLIENT_MAX_IDLE_CONNS" envDefault:"100"`
	MaxConnsPerHost int           `env:"HTTP_CLIENT_MAX_CONNS_PER_HOST" envDefault:"10"`
}

// Client is a thin wrapper around *http.Client with convenience Do* methods.
type Client struct {
	inner *http.Client
}

func NewClient(cfg ClientConfig) *Client {
	transport := &http.Transport{
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxConnsPerHost: cfg.MaxConnsPerHost,
	}
	zap.L().Info("http client created", zap.Duration("timeout", cfg.Timeout))
	return &Client{
		inner: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
	}
}

// do is the single entry point for all requests.
func (c *Client) do(ctx context.Context, method, rawURL string, body io.Reader, ct ContentType, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("httpclient: build request: %w", err)
	}

	if ct != "" {
		req.Header.Set("Content-Type", string(ct))
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := c.inner.Do(req)
	log := logger.FromContext(ctx)
	if err != nil {
		log.Error("http request failed",
			zap.String("method", method),
			zap.String("url", rawURL),
			zap.Duration("latency", time.Since(start)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("httpclient: %w", err)
	}
	log.Info("http request",
		zap.String("method", method),
		zap.String("url", rawURL),
		zap.Int("status", resp.StatusCode),
		zap.Duration("latency", time.Since(start)),
	)
	return resp, nil
}

func encodeJSON(body any) (io.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("httpclient: marshal json: %w", err)
	}
	return bytes.NewReader(b), nil
}

// ── No-body methods ───────────────────────────────────────────────────────────

func (c *Client) DoGet(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, rawURL, nil, "", headers)
}

func (c *Client) DoDelete(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, rawURL, nil, "", headers)
}

// ── JSON body methods ─────────────────────────────────────────────────────────

func (c *Client) DoPostJSON(ctx context.Context, rawURL string, body any, headers map[string]string) (*http.Response, error) {
	r, err := encodeJSON(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPost, rawURL, r, ContentTypeJSON, headers)
}

func (c *Client) DoPutJSON(ctx context.Context, rawURL string, body any, headers map[string]string) (*http.Response, error) {
	r, err := encodeJSON(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPut, rawURL, r, ContentTypeJSON, headers)
}

func (c *Client) DoPatchJSON(ctx context.Context, rawURL string, body any, headers map[string]string) (*http.Response, error) {
	r, err := encodeJSON(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPatch, rawURL, r, ContentTypeJSON, headers)
}

// ── Form body methods ─────────────────────────────────────────────────────────

func (c *Client) DoPostForm(ctx context.Context, rawURL string, body url.Values, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, rawURL, strings.NewReader(body.Encode()), ContentTypeForm, headers)
}

func (c *Client) DoPutForm(ctx context.Context, rawURL string, body url.Values, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodPut, rawURL, strings.NewReader(body.Encode()), ContentTypeForm, headers)
}

func (c *Client) DoPatchForm(ctx context.Context, rawURL string, body url.Values, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodPatch, rawURL, strings.NewReader(body.Encode()), ContentTypeForm, headers)
}
