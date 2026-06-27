package product

import (
	"context"

	"go.uber.org/zap"

	"github.com/your-org/service-name/internal/domain/entity"
	"github.com/your-org/service-name/internal/domain/port"
	"github.com/your-org/service-name/pkg/logger"
)

type Usecase struct {
	repo port.ProductRepository
}

func New(repo port.ProductRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) CreateProduct(ctx context.Context, product *entity.Product) error {
	if err := u.repo.Create(ctx, product); err != nil {
		logger.FromContext(ctx).Error("create product failed", zap.Error(err))
		return err
	}
	return nil
}

func (u *Usecase) GetProduct(ctx context.Context, id int64) (*entity.Product, error) {
	p, err := u.repo.GetByID(ctx, id)
	if err != nil {
		logger.FromContext(ctx).Error("get product failed", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}
	return p, nil
}

func (u *Usecase) ListProducts(ctx context.Context, offset, limit int) ([]*entity.Product, error) {
	products, err := u.repo.List(ctx, offset, limit)
	if err != nil {
		logger.FromContext(ctx).Error("list products failed", zap.Error(err))
		return nil, err
	}
	return products, nil
}

func (u *Usecase) UpdateProduct(ctx context.Context, product *entity.Product) error {
	if err := u.repo.Update(ctx, product); err != nil {
		logger.FromContext(ctx).Error("update product failed", zap.Int64("id", product.ID), zap.Error(err))
		return err
	}
	return nil
}

func (u *Usecase) DeleteProduct(ctx context.Context, id int64) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		logger.FromContext(ctx).Error("delete product failed", zap.Int64("id", id), zap.Error(err))
		return err
	}
	return nil
}
