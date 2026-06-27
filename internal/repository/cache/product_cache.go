package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/your-org/service-name/internal/consts"
	"github.com/your-org/service-name/internal/domain/entity"
)

const productTTL = 10 * time.Minute

type ProductCache struct {
	client *redis.Client
}

func New(client *redis.Client) *ProductCache {
	return &ProductCache{client: client}
}

func (c *ProductCache) Get(ctx context.Context, id int64) (*entity.Product, error) {
	data, err := c.client.Get(ctx, key(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var product entity.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (c *ProductCache) Set(ctx context.Context, product *entity.Product) error {
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key(product.ID), data, productTTL).Err()
}

func (c *ProductCache) Delete(ctx context.Context, id int64) error {
	return c.client.Del(ctx, key(id)).Err()
}

func key(id int64) string {
	return fmt.Sprintf("%s:%d", consts.CacheKeyPrefixProduct, id)
}
