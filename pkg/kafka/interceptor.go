package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/your-org/service-name/pkg/logger"
)

// Interceptor allows hooking before/after every Kafka record send or receive.
type Interceptor interface {
	Before(ctx context.Context, rec *kgo.Record) error
	After(ctx context.Context, rec *kgo.Record, err error)
}

type interceptorChain struct {
	interceptors []Interceptor
}

func newChain(interceptors []Interceptor) *interceptorChain {
	return &interceptorChain{interceptors: interceptors}
}

func (c *interceptorChain) Before(ctx context.Context, rec *kgo.Record) error {
	for _, i := range c.interceptors {
		if err := i.Before(ctx, rec); err != nil {
			logger.FromContext(ctx).Error("kafka interceptor before failed",
				zap.String("topic", rec.Topic),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func (c *interceptorChain) After(ctx context.Context, rec *kgo.Record, err error) {
	for _, i := range c.interceptors {
		i.After(ctx, rec, err)
	}
}
