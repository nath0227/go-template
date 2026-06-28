package kafka

import (
	"context"
	"encoding/json"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/your-org/service-name/internal/domain/entity"
	"github.com/your-org/service-name/internal/usecase/product"
	"github.com/your-org/service-name/pkg/logger"
)

type ProductConsumer struct {
	uc *product.Usecase
}

func NewProductConsumer(uc *product.Usecase) *ProductConsumer {
	return &ProductConsumer{uc: uc}
}

func (c *ProductConsumer) Handle(ctx context.Context, rec *kgo.Record) error {
	log := logger.FromContext(ctx)
	var p entity.Product
	if err := json.Unmarshal(rec.Value, &p); err != nil {
		log.Error("product consumer: unmarshal failed", zap.Error(err))
		return err
	}
	return c.uc.CreateProduct(ctx, &p)
}
