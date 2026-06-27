# Conventions

## Trace ID (TID) propagation

Every request carries a trace ID that flows through the entire call chain. The ID is a UUID v4 or an OpenTelemetry trace ID (if an OTel span is active).

### HTTP path

1. `middleware.TID()` runs first on every request.
2. It reads `X-Trace-ID` from the request header; if absent, calls `logger.NewTID(ctx)`.
3. Stores the ID in ctx via `logger.WithTID(ctx, tid)`.
4. Writes `X-Trace-ID` back in the response header.

### Kafka consumer path

`Consumer.Run()` calls `injectTIDFromHeaders(ctx, msg.Headers)` before invoking the handler. It reads the `"tid"` Kafka header and calls `logger.WithTID`. If the header is absent the ctx TID remains whatever the consumer's parent context had.

### Kafka producer path

When publishing events that should carry the calling request's TID, add the header manually. Use `logger.TIDKey` for the header name — do not write the string `"tid"` directly:

```go
msgs := []kafka.Message{{
    Value: body,
    Headers: []kafka.Header{{Key: logger.TIDKey, Value: []byte(logger.TIDFromContext(ctx))}},
}}
producer.WriteMessages(ctx, msgs...)
```

---

## Logger usage

**Always use `logger.FromContext(ctx)` in application code, never `zap.L()` directly.**

`logger.FromContext(ctx)` returns `zap.L().With("tid", <tid>)` when a TID is present, ensuring every log line is traceable.

```go
// Correct
log := logger.FromContext(ctx)
log.Error("something failed", zap.Error(err))

// Wrong — loses the tid field
zap.L().Error("something failed", zap.Error(err))
```

`zap.L()` directly is only acceptable in `pkg/` constructors (during startup, before any request context exists) — and those constructors already do this correctly.

---

## Configuration pattern

Each `pkg/` package owns its config struct with `env:` and `envDefault:` tags from `caarlos0/env/v11`. `config.Config` in `config/config.go` aggregates all of them.

```go
// Each package defines its own
type RedisConfig struct {
    Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
    Password string `env:"REDIS_PASSWORD"`
    ...
}

// config.go aggregates
type Config struct {
    DB      pkgdb.DBConfig
    Cache   cache.RedisConfig
    Kafka   kafka.KafkaConfig
    ...
}

// Load reads config/.env then env vars
cfg, err := config.Load()
```

`config.Load()` calls `godotenv.Load("config/.env")` silently (missing file is not an error), then `env.Parse`. Env vars in the shell always override the `.env` file.

**Do not add global variables or `init()` functions for configuration.** Always pass config structs through constructors.

---

## Error handling

- Errors are always propagated upward with `fmt.Errorf("context: %w", err)` where context helps.
- Usecase methods log errors with `logger.FromContext(ctx).Error(...)` and return the error.
- Handler methods map errors to HTTP status codes: repository-not-found errors → 404, all other errors → 500.
- The repository layer does not log — it returns errors and lets the caller decide.
- `pkg/` constructors return errors on failure; they never `log.Fatal`.

---

## Repository read-through cache pattern

`internal/repository/db/product_repo.go` wraps both GORM and the Redis cache:

```
GetByID:
  cache.Get → hit → return
           → miss → db.First → cache.Set → return

Create / Update:
  db operation → cache.Set (best-effort, error ignored with _)

Delete:
  db operation → cache.Delete (best-effort)
```

Cache key format: `product:<id>`. TTL: 10 minutes. Encoding: JSON.
Cache errors on write are silently ignored (`_ = r.cache.Set(...)`) — the DB is the source of truth.

---

## Routing convention

All HTTP routes are defined in **`internal/router/router.go`** only. Handlers expose plain methods (`Create`, `GetByID`, etc.) — they have no knowledge of routes.

When adding a new handler, register its routes in `router.New()`:

```go
func New(product *httphandler.ProductHandler, order *orderhandler.OrderHandler) *gin.Engine {
    r := gin.New()
    r.Use(middleware.Recovery(), middleware.TID(), middleware.Logger(), middleware.CORS())

    v1 := r.Group("/api/v1")
    {
        products := v1.Group("/products")
        products.POST("", product.Create)
        // ...

        orders := v1.Group("/orders")
        orders.POST("", order.Create)
        // ...
    }
    return r
}
```

---

## No `init()` functions

The codebase uses no `init()` functions. All initialization happens explicitly in `main.go` in a defined order. This makes the startup sequence readable and prevents implicit ordering bugs.

---

## String constants

**Rule: every string value used in code must be a named constant. Never write a bare string literal for any identifier, key, prefix, or message that will be used in logic.**

This applies to cache key prefixes, Kafka header names, HTTP header names, error message strings, and any other string that carries meaning beyond a one-off log message. The placement of the constant depends on which layer uses it.

### `pkg/` strings — export from the owning package

`pkg/` packages cannot import `internal/`, so strings shared across `pkg/` packages must be exported from whichever `pkg/` package owns the concept.

The `"tid"` string is owned by `pkg/logger` and exported as `logger.TIDKey`. It is used:
- as the context map key value (`tidCtxKey contextKey = TIDKey`)
- as the Zap field name (`zap.String(TIDKey, tid)`)
- as the Kafka message header key (`h.Key == logger.TIDKey` in `pkg/kafka/consumer.go`)

**Rule:** never write the bare string `"tid"` anywhere. Always reference `logger.TIDKey`.

### `internal/` strings — `internal/consts/consts.go`

Domain-scoped strings that are used across multiple `internal/` packages, or that benefit from central visibility (e.g. all cache key namespaces in one place), live in `internal/consts/consts.go`.

Current constants:

```go
package consts

const (
    CacheKeyPrefixProduct = "product"   // Redis key: "product:<id>"
    ErrMsgInvalidID       = "invalid id" // HTTP 400 response body
)
```

When adding a new entity, add its cache prefix here:

```go
CacheKeyPrefixOrder = "order"   // Redis key: "order:<id>"
CacheKeyPrefixUser  = "user"
```

This prevents cache key collisions between resources and makes all namespaces visible in one place.

### Exceptions — what does NOT need a constant

Only two categories are exempt:

- **Log message strings** — human-readable descriptions like `"gorm connected"` or `"kafka producer created"`. These are not identifiers and are never compared or matched in logic.
- **Framework-convention keys** — e.g. `"error"` in `gin.H{"error": ...}`. This is a REST/Gin convention enforced by documentation, not a project string.

Everything else — cache key prefixes, Kafka header names, HTTP header names, error response messages, Redis key formats — must be a named constant.
