# Appendix — Environment Variable Reference

All variables with defaults are optional. Variables with no default are required if the corresponding feature is used.

---

## Server

| Variable | Default | Notes |
|---|---|---|
| `SERVER_PORT` | `8080` | |
| `SERVER_READ_TIMEOUT` | `30s` | |
| `SERVER_WRITE_TIMEOUT` | `30s` | |
| `SERVER_SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown window |

---

## Logger

| Variable | Default | Notes |
|---|---|---|
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` (production) or `console` (development) |

---

## Database (GORM)

| Variable | Default | Notes |
|---|---|---|
| `DB_DRIVER` | `postgres` | `postgres` or `mysql` |
| `DB_HOST` | `localhost:5432` | Includes port — no separate `DB_PORT` |
| `DB_USER` | — | |
| `DB_PASSWORD` | — | |
| `DB_NAME` | — | |
| `DB_SSL_MODE` | `disable` | Postgres: `disable`, `require`, `verify-full` |
| `DB_TIMEZONE` | `UTC` | MySQL only |
| `DB_MAX_OPEN_CONNS` | `25` | |
| `DB_MAX_IDLE_CONNS` | `5` | |
| `DB_CONN_MAX_LIFETIME` | `5m` | Go duration string |

---

## Redis

| Variable | Default | Notes |
|---|---|---|
| `REDIS_ADDR` | `localhost:6379` | `host:port` |
| `REDIS_PASSWORD` | `""` | |
| `REDIS_DB` | `0` | |
| `REDIS_POOL_SIZE` | `10` | |

---

## Kafka

| Variable | Default | Notes |
|---|---|---|
| `KAFKA_BROKERS` | — | Comma-separated list, e.g. `b1:9092,b2:9092` |
| `KAFKA_AUTH` | `none` | `none`, `sasl_scram_512`, `aws_msk` |
| `KAFKA_USERNAME` | — | Required for `sasl_scram_512` |
| `KAFKA_PASSWORD` | — | Required for `sasl_scram_512` |
| `KAFKA_AWS_REGION` | — | Required for `aws_msk` |
| `KAFKA_PRODUCER_TOPIC` | — | |
| `KAFKA_PRODUCER_BATCH_SIZE` | `100` | |
| `KAFKA_PRODUCER_BATCH_TIMEOUT` | `1ms` | |
| `KAFKA_CONSUMER_TOPIC` | — | |
| `KAFKA_CONSUMER_GROUP_ID` | — | |
| `KAFKA_CONSUMER_MIN_BYTES` | `1` | |
| `KAFKA_CONSUMER_MAX_BYTES` | `10485760` | 10 MiB |
| `KAFKA_CONSUMER_COMMIT_INTERVAL` | `1s` | |

---

## Object Storage

| Variable | Default | Notes |
|---|---|---|
| `STORAGE_DRIVER` | `aws` | `aws`, `gcp`, or `tencent` |
| `STORAGE_REGION` | — | AWS / GCP region |
| `STORAGE_BUCKET` | — | Default bucket name |
| `STORAGE_ACCESS_KEY_ID` | — | AWS / MinIO / Tencent |
| `STORAGE_SECRET_ACCESS_KEY` | — | AWS / MinIO / Tencent |
| `STORAGE_ENDPOINT` | `""` | Override endpoint — use `http://localhost:9000` for MinIO |
| `STORAGE_CREDENTIALS_FILE` | `""` | GCP service account JSON path |
| `STORAGE_CDN_BASE_URL` | `""` | If set, `PublicURL` returns `<CDN_BASE_URL>/<key>` |

---

## HTTP Client

| Variable | Default | Notes |
|---|---|---|
| `HTTP_CLIENT_TIMEOUT` | `30s` | Per-request timeout |
| `HTTP_CLIENT_MAX_IDLE_CONNS` | `100` | Transport pool |
| `HTTP_CLIENT_MAX_CONNS_PER_HOST` | `10` | |

---

## Config structs quick reference

```go
// config/config.go
type Config struct {
    DB      pkgdb.DBConfig
    Cache   cache.RedisConfig
    Kafka   kafka.KafkaConfig
    Storage storage.StorageConfig
    Client  httpclient.ClientConfig
    Server  httpserver.ServerConfig
    Log     logger.LogConfig
}
```

Load order: `config/.env` file (if present) → environment variables. Env vars always win.

---

## docker-compose services

| Service | Image | Ports | Notes |
|---|---|---|---|
| `postgres` | `postgres:16-alpine` | `5432` | Health-checked |
| `redis` | `redis:7-alpine` | `6379` | Health-checked |
| `zookeeper` | `confluentinc/cp-zookeeper:7.6.0` | `2181` | |
| `kafka` | `confluentinc/cp-kafka:7.6.0` | `9092` | |
| `minio` | `minio/minio:latest` | `9000` (API), `9001` (console) | Health-checked; `minioadmin`/`minioadmin` |

`app` service waits for postgres, redis, and minio healthy before starting.

---

## Key dependencies and their roles

| Package | Version | Role |
|---|---|---|
| `gin-gonic/gin` | v1.10.0 | HTTP router and middleware |
| `segmentio/kafka-go` | v0.4.47 | Kafka producer/consumer |
| `gorm.io/gorm` | v1.25.10 | ORM |
| `jackc/pgx/v5` | v5.5.5 | Raw Postgres pool (optional) |
| `redis/go-redis/v9` | v9.5.1 | Redis client |
| `aws-sdk-go-v2` | v1.32.4 | S3 + MinIO storage |
| `aws-msk-iam-sasl-signer-go` | v1.0.4 | AWS MSK IAM auth |
| `caarlos0/env/v11` | v11.1.0 | Struct-tag env parsing |
| `joho/godotenv` | v1.5.1 | `.env` file loading |
| `go.uber.org/zap` | v1.27.0 | Structured logging |
| `google/uuid` | v1.6.0 | TID generation fallback |
| `otel/trace` | v1.24.0 | OTel span extraction for TID |
