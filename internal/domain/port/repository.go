package port

import (
	"context"

	"github.com/your-org/service-name/internal/domain/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id int64) (*entity.Product, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id int64) error
}
