# pkg/ Packages

Each `pkg/` package is self-contained. None import from `internal/`. All constructors return an error when a connection cannot be established.

---

## pkg/logger

**`logger.Init(cfg LogConfig) error`**
Builds a Zap logger and calls `zap.ReplaceGlobals`. Must be called first in `main`, before any other constructor (they all log on startup).

Format `"json"` → production config. Format `"console"` → development config with colours.

**`logger.FromContext(ctx) *zap.Logger`**
Returns `zap.L().With("tid", <tid>)` if a TID is in ctx, otherwise `zap.L()`.
Use this everywhere in application code instead of `zap.L()` directly.

**`logger.WithTID(ctx, tid) context.Context`** / **`logger.TIDFromContext(ctx) string`**
Store and retrieve a trace ID in a context. The key is an unexported `contextKey("tid")`.

**`logger.NewTID(ctx) string`**
Generates a trace ID. If an active OTel span exists with a valid trace ID it returns that; otherwise generates a UUID v4. Used by the TID middleware.

---

## pkg/db

**`db.NewGORM(cfg DBConfig) (*gorm.DB, error)`**
Opens a GORM connection, sets connection-pool parameters, pings with `PingContext`, then logs "gorm connected". Supports `DB_DRIVER=postgres` (default) and `mysql`.

**`db.NewPGX(cfg DBConfig) (*pgxpool.Pool, error)`**
Optional raw pgx pool for queries that bypass GORM. Pings on startup.

`DB_HOST` carries `host:port` (e.g. `localhost:5432`). There is no separate `DB_PORT`.

DSN formats built internally:
- Postgres: `postgres://user:pass@host:port/db?sslmode=...`
- MySQL: `user:pass@tcp(host:port)/db?parseTime=True&loc=...`

---

## pkg/cache

**`cache.NewRedis(cfg RedisConfig) (*redis.Client, error)`**
Creates and pings a go-redis client. The `*redis.Client` is passed directly to `repository/cache/product_cache.go`.

---

## pkg/kafka

Single `KafkaConfig` struct covers both producer and consumer. Auth is selected by `KAFKA_AUTH`:

| Value | Mechanism |
|---|---|
| `none` | plaintext (default) |
| `sasl_scram_512` | SASL SCRAM-SHA-512 via `kafka-go/sasl/scram` |
| `aws_msk` | OAUTHBEARER via `aws-msk-iam-sasl-signer-go`; token refreshed per connection in `mskMechanism.Start()` |

**`kafka.NewProducer(cfg, interceptors...) (*Producer, error)`**
Wraps `kafka.Writer`. Call `producer.WriteMessages(ctx, msgs...)`. Interceptors run `Before` each batch and `After` — if any `Before` returns an error the write is skipped.

**`kafka.NewConsumer(cfg, interceptors...) (*Consumer, error)`**
Wraps `kafka.Reader`. Call `consumer.Run(ctx, handler)` in a goroutine. It loops forever fetching messages, extracting TID from the `"tid"` Kafka header, running interceptors, calling the handler, and committing on success. Returns when `ctx` is cancelled.

**`kafka.Interceptor` interface**
```go
type Interceptor interface {
    Before(ctx context.Context, msg *kafka.Message) error
    After(ctx context.Context, msg *kafka.Message, err error)
}
```
Chain stops at the first `Before` error. `After` is always called on all interceptors.

---

## pkg/storage

**`storage.NewStorage(cfg StorageConfig) (ObjectStorage, error)`**
Factory selecting driver by `STORAGE_DRIVER`: `aws`, `gcp`, or `tencent`.

```go
type ObjectStorage interface {
    Upload(ctx context.Context, bucket, key string, r io.Reader) error
    Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, bucket, key string) error
    PublicURL(bucket, key string) string
}
```

All three implementations (`s3.go`, `gcs.go`, `cos.go`) are in `package storage` (flat) to avoid circular imports — they all depend on `StorageConfig`.

**MinIO** works with the `aws` driver. Set `STORAGE_ENDPOINT=http://localhost:9000`; the S3 client automatically enables path-style addressing when an endpoint is provided.

`PublicURL` returns a CDN URL if `STORAGE_CDN_BASE_URL` is set; otherwise falls back to the default cloud URL.

---

## pkg/httpclient

**`httpclient.NewClient(cfg ClientConfig) *Client`**
Wraps `*http.Client` with a connection-pooled transport. All methods call the shared `do()` helper which sets `Content-Type`, applies extra headers, and logs method/url/status/latency via `logger.FromContext(ctx)`.

```go
// No body
client.DoGet(ctx, url, headers)
client.DoDelete(ctx, url, headers)

// JSON body — marshals `body any` with json.Marshal, sets Content-Type: application/json
client.DoPostJSON(ctx, url, body, headers)
client.DoPutJSON(ctx, url, body, headers)
client.DoPatchJSON(ctx, url, body, headers)

// Form body — encodes url.Values, sets Content-Type: application/x-www-form-urlencoded
client.DoPostForm(ctx, url, values, headers)
client.DoPutForm(ctx, url, values, headers)
client.DoPatchForm(ctx, url, values, headers)
```

All methods return `(*http.Response, error)`. Non-2xx responses are NOT errors — the caller must check `resp.StatusCode`.

---

## pkg/httpserver

**`httpserver.NewServer(cfg ServerConfig, router *gin.Engine) *Server`**
**`server.Start()`**
Starts listening in a goroutine, waits for `SIGINT`/`SIGTERM`, then gracefully shuts down within `ShutdownTimeout`.

### Middleware (pkg/httpserver/middleware)

Applied in `internal/router/router.go` in this order: `Recovery → TID → Logger → CORS`.

| Middleware | Behaviour |
|---|---|
| `Recovery()` | Catches panics, logs with `zap.L().Error`, returns 500 |
| `TID()` | Reads `X-Trace-ID` request header or generates one via `logger.NewTID`; injects into ctx via `logger.WithTID`; echoes in response header |
| `Logger()` | Logs method, path, status, latency via `logger.FromContext(c.Request.Context())` |
| `CORS()` | Sets permissive CORS headers; short-circuits OPTIONS with 204 |
