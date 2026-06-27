# Architecture

## Layer diagram

```
┌─────────────────────────────────────────────────────┐
│  cmd/server/main.go  (wiring only)                  │
└─────────────────────┬───────────────────────────────┘
                      │
        ┌─────────────▼─────────────┐
        │  internal/router          │  ← all HTTP routes
        └─────────────┬─────────────┘
                      │
        ┌─────────────▼─────────────────────────────┐
        │  internal/handler/http    (Gin handlers)   │
        │  internal/handler/kafka   (message handler)│
        └─────────────┬─────────────────────────────┘
                      │
        ┌─────────────▼─────────────┐
        │  internal/usecase/product │  ← business logic
        └─────────────┬─────────────┘
                      │ (depends on port interface)
        ┌─────────────▼─────────────────────────────┐
        │  internal/domain/port     (interfaces)     │
        │  internal/domain/entity   (plain structs)  │
        └─────────────┬─────────────────────────────┘
                      │ (implemented by)
        ┌─────────────▼─────────────────────────────┐
        │  internal/repository/db   (GORM + cache)  │
        │  internal/repository/cache (Redis)         │
        └─────────────┬─────────────────────────────┘
                      │
        ┌─────────────▼─────────────┐
        │  pkg/  (infra packages)   │
        └───────────────────────────┘
```

---

## Dependency rules

These rules are strict. Violating them creates circular imports or breaks the copy-and-own promise.

1. **`pkg/` packages never import from `internal/`.**
   They are self-contained infrastructure wrappers. Any project can copy a `pkg/` package out.

2. **`internal/domain/` has no imports from the rest of `internal/`.**
   Entities and port interfaces are the dependency anchor — everything else points at them.

3. **`internal/usecase/` depends only on `internal/domain/port` interfaces, never on concrete repo types.**
   This is what makes usecases unit-testable without a real DB.

4. **`internal/repository/` imports `internal/domain/entity` and `pkg/` — never `internal/usecase/` or `internal/handler/`.**

5. **`internal/handler/` imports `internal/usecase/` (concrete types) — never `internal/repository/` directly.**

6. **`internal/router/` imports `internal/handler/http` and `pkg/httpserver/middleware` — nothing else from `internal/`.**

---

## Data flow — HTTP request

```
Client
  │  POST /api/v1/products  (body: JSON Product)
  │  X-Trace-ID: <optional>
  ▼
middleware.TID()           ← reads/generates trace ID, injects into ctx, echoes header
middleware.Logger()        ← logs method, path, status, latency via logger.FromContext(ctx)
middleware.Recovery()      ← catches panics → 500
  ▼
ProductHandler.Create()    ← ShouldBindJSON → calls usecase
  ▼
product.Usecase.CreateProduct()   ← calls repo port interface
  ▼
ProductRepo.Create()       ← GORM insert → cache.Set()
  ▼
Response 201 JSON
```

## Data flow — Kafka consumer

```
Kafka broker → Consumer.Run()
  │  FetchMessage
  │  injectTIDFromHeaders()   ← reads "tid" header → logger.WithTID(ctx)
  │  chain.Before()           ← interceptors
  │  handler(ctx, msg)        ← ProductConsumer.Handle()
  │    json.Unmarshal → Usecase.CreateProduct()
  │  chain.After()
  │  CommitMessages (only on success)
```

---

## File tree

```
code/
├── cmd/server/main.go
├── config/
│   ├── config.go           ← Config struct aggregating all pkg configs
│   ├── config.yaml         ← (unused by code; kept for reference)
│   └── .env.example
├── internal/
│   ├── consts/consts.go             ← shared string constants (cache prefixes, error msgs)
│   ├── domain/
│   │   ├── entity/product.go
│   │   └── port/
│   │       ├── repository.go   ← ProductRepository interface
│   │       └── service.go      ← example external service interface
│   ├── handler/
│   │   ├── http/product_handler.go
│   │   └── kafka/product_consumer.go
│   ├── repository/
│   │   ├── db/product_repo.go       ← GORM + read-through cache
│   │   └── cache/product_cache.go   ← Redis, key "product:<id>", TTL 10m
│   ├── router/router.go             ← ALL routes + middleware setup
│   └── usecase/product/usecase.go
├── pkg/
│   ├── cache/redis.go
│   ├── db/
│   │   ├── gorm.go     ← NewGORM (postgres + mysql)
│   │   └── pgx.go      ← NewPGX (raw pool, optional)
│   ├── httpclient/client.go
│   ├── httpserver/
│   │   ├── server.go
│   │   └── middleware/
│   │       ├── cors.go
│   │       ├── logger.go
│   │       ├── recovery.go
│   │       └── tid.go
│   ├── kafka/
│   │   ├── auth.go
│   │   ├── config.go
│   │   ├── consumer.go
│   │   ├── interceptor.go
│   │   └── producer.go
│   ├── logger/
│   │   ├── context.go
│   │   └── logger.go
│   └── storage/
│       ├── storage.go   ← interface + factory
│       ├── s3.go
│       ├── gcs.go
│       └── cos.go
├── docs/ai/             ← this directory
├── migrations/
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── go.mod
```
