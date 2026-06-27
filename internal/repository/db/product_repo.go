package db

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/your-org/service-name/internal/domain"
	"github.com/your-org/service-name/internal/domain/entity"
	cacherepo "github.com/your-org/service-name/internal/repository/cache"
	"github.com/your-org/service-name/pkg/storage"
)

type ProductRepo struct {
	db    *gorm.DB
	cache *cacherepo.ProductCache
	store storage.ObjectStorage
}

func NewProductRepo(db *gorm.DB, rdb *redis.Client, store storage.ObjectStorage) *ProductRepo {
	return &ProductRepo{
		db:    db,
		cache: cacherepo.New(rdb),
		store: store,
	}
}

func (r *ProductRepo) Create(ctx context.Context, product *entity.Product) error {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return err
	}
	_ = r.cache.Set(ctx, product)
	return nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	if p, err := r.cache.Get(ctx, id); err == nil {
		return p, nil
	}

	var product entity.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	_ = r.cache.Set(ctx, &product)
	return &product, nil
}

func (r *ProductRepo) List(ctx context.Context, offset, limit int) ([]*entity.Product, error) {
	var products []*entity.Product
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepo) Update(ctx context.Context, product *entity.Product) error {
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		return err
	}
	_ = r.cache.Set(ctx, product)
	return nil
}

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Product{}, id).Error; err != nil {
		return err
	}
	_ = r.cache.Delete(ctx, id)
	return nil
}
