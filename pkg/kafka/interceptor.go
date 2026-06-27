package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/your-org/service-name/pkg/logger"
)

// Interceptor allows hooking before/after every Kafka message send or receive.
type Interceptor interface {
	Before(ctx context.Context, msg *kafka.Message) error
	After(ctx context.Context, msg *kafka.Message, err error)
}

type interceptorChain struct {
	interceptors []Interceptor
}

func newChain(interceptors []Interceptor) *interceptorChain {
	return &interceptorChain{interceptors: interceptors}
}

func (c *interceptorChain) Before(ctx context.Context, msg *kafka.Message) error {
	for _, i := range c.interceptors {
		if err := i.Before(ctx, msg); err != nil {
			logger.FromContext(ctx).Error("kafka interceptor before failed",
				zap.String("topic", msg.Topic),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func (c *interceptorChain) After(ctx context.Context, msg *kafka.Message, err error) {
	for _, i := range c.interceptors {
		i.After(ctx, msg, err)
	}
}
