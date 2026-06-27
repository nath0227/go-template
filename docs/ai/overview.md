# Overview

## What this is

A copy-and-own Go microservice boilerplate. It is a single-service template — not a mono-repo framework. Clone it once per service and rename everything to suit.

It ships with:
- HTTP API (Gin)
- Kafka producer + consumer
- PostgreSQL via GORM (also ships a raw pgx pool constructor)
- Redis cache
- Multi-cloud object storage (AWS S3 / GCP GCS / Tencent COS / MinIO)
- Outbound HTTP client
- Structured JSON logging (Zap) with trace-ID propagation

The `Product` domain is a placeholder. Delete it and replace with your own domain.

---

## Module path

**`go.mod` declares `module github.com/your-org/service-name`.**

This must be the first thing you change. Replace every occurrence throughout the codebase with your real module path, then run `go mod tidy`.

```bash
# macOS / Linux
grep -r "github.com/your-org/service-name" --include="*.go" -l | \
  xargs sed -i '' 's|github.com/your-org/service-name|github.com/acme/my-service|g'
# Also update go.mod module line
sed -i '' 's|github.com/your-org/service-name|github.com/acme/my-service|g' go.mod
go mod tidy
```

---

## Tech stack

| Concern | Library |
|---|---|
| HTTP framework | `github.com/gin-gonic/gin` |
| Kafka | `github.com/segmentio/kafka-go` |
| ORM | `gorm.io/gorm` + `gorm.io/driver/postgres` / `mysql` |
| Raw PG pool | `github.com/jackc/pgx/v5/pgxpool` |
| Redis | `github.com/redis/go-redis/v9` |
| Object storage | `aws-sdk-go-v2`, `cloud.google.com/go/storage`, `tencentyun/cos-go-sdk-v5` |
| Logger | `go.uber.org/zap` |
| Config | `github.com/caarlos0/env/v11` + `github.com/joho/godotenv` |
| Trace ID | `go.opentelemetry.io/otel/trace` + `github.com/google/uuid` |
| Go version | 1.25 |

---

## Startup sequence (`cmd/server/main.go`)

The `main` function is wiring only — no logic lives here.

```
config.Load()
  └── logger.Init()
        └── pkgdb.NewGORM()          ← pings DB before logging "connected"
              └── pkgcache.NewRedis()
                    └── pkgstorage.NewStorage()
                          └── pkgkafka.NewProducer()
                                └── pkgkafka.NewConsumer()
                                      └── dbrepo.NewProductRepo()
                                            └── productusecase.New()
                                                  └── httphandler.NewProductHandler()
                                                        └── router.New()
                                                              └── httpserver.NewServer().Start()
                                                                    └── go consumer.Run()
```

`server.Start()` blocks until `SIGINT` or `SIGTERM`, then gracefully shuts down HTTP. Consumer stops when its context is cancelled (the `defer cancel()` in main).

---

## Local development

```bash
docker-compose up -d          # starts postgres, redis, kafka, minio
cp config/.env.example config/.env
go run ./cmd/server
```

MinIO web console is available at http://localhost:9001 (user: `minioadmin`, password: `minioadmin`).
