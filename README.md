# go-template

A copy-and-own Go microservice boilerplate. Clone once per service, rename the module, and start building your domain.

## What's included

| Concern | Library |
|---|---|
| HTTP framework | Gin |
| Kafka | segmentio/kafka-go |
| ORM | GORM (PostgreSQL / MySQL) + raw pgx pool |
| Cache | Redis (go-redis v9) |
| Object storage | AWS S3 · GCP GCS · Tencent COS · MinIO |
| Logger | Zap (structured JSON, trace-ID propagation) |
| Config | caarlos0/env + godotenv |
| Go version | 1.25 |

## Project layout

```
cmd/server/          ← entry point (wiring only, no logic)
config/              ← config struct + .env loader
internal/
  domain/            ← entities and port interfaces (no external deps)
  usecase/           ← business logic (depends on ports)
  repository/        ← DB + cache implementations
  handler/           ← HTTP (Gin) and Kafka consumers
  router/            ← all HTTP routes
  consts/            ← shared string constants
pkg/                 ← self-contained infra packages
  db/                ← GORM + pgx constructors
  cache/             ← Redis client
  kafka/             ← producer + consumer + IAM auth
  storage/           ← multi-cloud object storage abstraction
  httpserver/        ← Gin server + middleware (CORS, logger, recovery, TID)
  httpclient/        ← outbound HTTP client
  logger/            ← Zap init + context helpers
migrations/          ← SQL migration files (golang-migrate)
docs/ai/             ← AI-assistant documentation
```

## Getting started

### 1. Rename the module

Replace every occurrence of `github.com/your-org/service-name` with your actual module path:

```bash
grep -r "github.com/your-org/service-name" --include="*.go" -l | \
  xargs sed -i '' 's|github.com/your-org/service-name|github.com/acme/my-service|g'
sed -i '' 's|github.com/your-org/service-name|github.com/acme/my-service|g' go.mod
go mod tidy
```

### 2. Start infrastructure

```bash
docker-compose up -d
```

Services started: PostgreSQL (5432), Redis (6379), Kafka (9092), MinIO S3 API (9000) + console (9001).

### 3. Configure

```bash
cp config/.env.example config/.env
# edit config/.env with your credentials
```

### 4. Run

```bash
make run
# or
go run ./cmd/server
```

Health check: `GET /system/health`

## API endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/products` | Create product |
| `GET` | `/api/v1/products` | List products |
| `GET` | `/api/v1/products/:id` | Get product by ID |
| `PUT` | `/api/v1/products/:id` | Update product |
| `DELETE` | `/api/v1/products/:id` | Delete product |

> The `Product` domain is a placeholder — delete it and replace with your own.

## Make targets

| Target | Description |
|---|---|
| `make build` | Compile binary to `bin/service-name` |
| `make run` | Build and run |
| `make test` | Run tests with race detector |
| `make lint` | Run golangci-lint |
| `make tidy` | `go mod tidy` |
| `make docker-up` | Start all infra via docker-compose |
| `make docker-down` | Stop infra |
| `make migrate` | Run DB migrations (requires `DB_DSN` env var) |

## Extending the template

See [`docs/ai/extending.md`](docs/ai/extending.md) for a step-by-step guide to adding a new domain resource end-to-end.
