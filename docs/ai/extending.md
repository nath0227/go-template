# Extending the Template

This guide shows exactly which files to create or modify when adding a new domain resource. The `Order` resource is used as the example throughout.

---

## Step 1 — Entity

**Create `internal/domain/entity/order.go`**

```go
package entity

import "time"

type Order struct {
    ID        int64     `json:"id" gorm:"primaryKey"`
    UserID    int64     `json:"user_id" gorm:"not null"`
    Total     float64   `json:"total" gorm:"not null"`
    Status    string    `json:"status" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

Entities are plain structs with JSON and GORM tags. No methods, no imports from `internal/`.

---

## Step 2 — Port interface

**Create `internal/domain/port/order_repository.go`**

```go
package port

import (
    "context"
    "github.com/your-org/service-name/internal/domain/entity"
)

type OrderRepository interface {
    Create(ctx context.Context, order *entity.Order) error
    GetByID(ctx context.Context, id int64) (*entity.Order, error)
    List(ctx context.Context, userID int64, offset, limit int) ([]*entity.Order, error)
    UpdateStatus(ctx context.Context, id int64, status string) error
    Delete(ctx context.Context, id int64) error
}
```

The interface lives in `port/`, not in the usecase or repository. Both the usecase and the repository implementation depend on this file.

---

## Step 3 — Cache (optional)

**Create `internal/repository/cache/order_cache.go`**

Follow the same pattern as `product_cache.go`: key format `order:<id>`, TTL, JSON encode/decode, `Get`/`Set`/`Delete` methods.

---

## Step 4 — Repository

**Create `internal/repository/db/order_repo.go`**

```go
package db

import (
    "context"
    "gorm.io/gorm"
    "github.com/redis/go-redis/v9"
    cacherepo "github.com/your-org/service-name/internal/repository/cache"
    "github.com/your-org/service-name/internal/domain/entity"
)

type OrderRepo struct {
    db    *gorm.DB
    cache *cacherepo.OrderCache
}

func NewOrderRepo(db *gorm.DB, rdb *redis.Client) *OrderRepo {
    return &OrderRepo{db: db, cache: cacherepo.NewOrderCache(rdb)}
}

// Implement all methods from port.OrderRepository
```

The repo implements `port.OrderRepository` but does **not** declare that it does (`implements` is implicit in Go). The usecase receives it as a `port.OrderRepository` interface.

---

## Step 5 — Usecase

**Create `internal/usecase/order/usecase.go`**

```go
package order

import (
    "context"
    "go.uber.org/zap"
    "github.com/your-org/service-name/internal/domain/entity"
    "github.com/your-org/service-name/internal/domain/port"
    "github.com/your-org/service-name/pkg/logger"
)

type Usecase struct {
    repo port.OrderRepository
}

func New(repo port.OrderRepository) *Usecase {
    return &Usecase{repo: repo}
}

func (u *Usecase) CreateOrder(ctx context.Context, order *entity.Order) error {
    if err := u.repo.Create(ctx, order); err != nil {
        logger.FromContext(ctx).Error("create order failed", zap.Error(err))
        return err
    }
    return nil
}
// ...
```

**Rule:** the usecase only takes `port.OrderRepository`, never the concrete `*db.OrderRepo`.

---

## Step 6 — HTTP handler

**Create `internal/handler/http/order_handler.go`**

```go
package http

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/your-org/service-name/internal/domain/entity"
    "github.com/your-org/service-name/internal/usecase/order"
)

type OrderHandler struct {
    uc *order.Usecase
}

func NewOrderHandler(uc *order.Usecase) *OrderHandler {
    return &OrderHandler{uc: uc}
}

func (h *OrderHandler) Create(c *gin.Context) { ... }
func (h *OrderHandler) GetByID(c *gin.Context) { ... }
// ...
```

Handlers do not define routes. They only receive a `*gin.Context` and call the usecase.

---

## Step 7 — Register routes

**Edit `internal/router/router.go`**

```go
func New(
    product *httphandler.ProductHandler,
    order   *orderhandler.OrderHandler,   // ← add parameter
) *gin.Engine {
    r := gin.New()
    r.Use(middleware.Recovery(), middleware.TID(), middleware.Logger(), middleware.CORS())

    v1 := r.Group("/api/v1")
    {
        products := v1.Group("/products")
        products.POST("", product.Create)
        // ...

        orders := v1.Group("/orders")          // ← add group
        orders.POST("", order.Create)
        orders.GET("/:id", order.GetByID)
        // ...
    }
    return r
}
```

---

## Step 8 — Wire in main.go

**Edit `cmd/server/main.go`**

```go
// After existing repo/usecase wiring:
orderRepo := dbrepo.NewOrderRepo(db, rdb)
orderUC   := orderusecase.New(orderRepo)

r := router.New(
    httphandler.NewProductHandler(uc),
    orderhandler.NewOrderHandler(orderUC),
)
```

---

## Step 9 — Kafka consumer (if needed)

**Create `internal/handler/kafka/order_consumer.go`**

Follow the same pattern as `product_consumer.go`. Add a new `KafkaConfig` consumer topic or reuse the existing consumer with a message-type discriminator.

Wire in `main.go`:

```go
go consumer.Run(ctx, kafkahandler.NewOrderConsumer(orderUC).Handle)
```

If you need separate topics for product and order, create two `Consumer` instances pointing at different `KAFKA_CONSUMER_TOPIC` values.

---

## Checklist

- [ ] `internal/consts/consts.go` — add `CacheKeyPrefix<Resource>` constant
- [ ] `internal/domain/entity/<resource>.go`
- [ ] `internal/domain/port/<resource>_repository.go`
- [ ] `internal/repository/cache/<resource>_cache.go` (optional) — use `consts.CacheKeyPrefix<Resource>`
- [ ] `internal/repository/db/<resource>_repo.go`
- [ ] `internal/usecase/<resource>/usecase.go`
- [ ] `internal/handler/http/<resource>_handler.go`
- [ ] Routes added to `internal/router/router.go`
- [ ] Wiring added to `cmd/server/main.go`
- [ ] DB migration added to `migrations/`
