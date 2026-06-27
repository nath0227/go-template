package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/your-org/service-name/pkg/logger"
)

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	reader *kafka.Reader
	chain  *interceptorChain
}

func NewConsumer(cfg KafkaConfig, interceptors ...Interceptor) (*Consumer, error) {
	dialer, err := buildDialer(cfg)
	if err != nil {
		return nil, err
	}

	rcfg := kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.ConsumerTopic,
		GroupID:        cfg.ConsumerGroupID,
		MinBytes:       cfg.ConsumerMinBytes,
		MaxBytes:       cfg.ConsumerMaxBytes,
		CommitInterval: cfg.ConsumerCommitInterval,
	}
	if dialer != nil {
		rcfg.Dialer = dialer
	}

	zap.L().Info("kafka consumer created",
		zap.String("topic", cfg.ConsumerTopic),
		zap.String("group", cfg.ConsumerGroupID),
		zap.String("auth", string(cfg.Auth)),
	)
	return &Consumer{reader: kafka.NewReader(rcfg), chain: newChain(interceptors)}, nil
}

func (c *Consumer) Run(ctx context.Context, handler MessageHandler) {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.FromContext(ctx).Error("kafka fetch message failed", zap.Error(err))
			continue
		}

		// Propagate tid from Kafka message headers into the context
		ctx = injectTIDFromHeaders(ctx, msg.Headers)

		if err := c.chain.Before(ctx, &msg); err != nil {
			continue
		}

		handlerErr := handler(ctx, msg)
		c.chain.After(ctx, &msg, handlerErr)

		if handlerErr != nil {
			logger.FromContext(ctx).Error("kafka message handler failed", zap.Error(handlerErr))
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			logger.FromContext(ctx).Error("kafka commit message failed", zap.Error(err))
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

// injectTIDFromHeaders reads the "tid" Kafka header and injects it into ctx.
func injectTIDFromHeaders(ctx context.Context, headers []kafka.Header) context.Context {
	for _, h := range headers {
		if h.Key == logger.TIDKey && len(h.Value) > 0 {
			return logger.WithTID(ctx, string(h.Value))
		}
	}
	return ctx
}
