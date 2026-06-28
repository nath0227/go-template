package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/your-org/service-name/pkg/logger"
)

type MessageHandler func(ctx context.Context, rec *kgo.Record) error

type Consumer struct {
	client *kgo.Client
	chain  *interceptorChain
}

func NewConsumer(cfg KafkaConfig, interceptors ...Interceptor) (*Consumer, error) {
	authOpts, err := buildAuthOpts(cfg)
	if err != nil {
		return nil, err
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumeTopics(cfg.ConsumerTopic),
		kgo.ConsumerGroup(cfg.ConsumerGroupID),
		kgo.Balancers(cfg.balancers()...),
		kgo.ConsumeResetOffset(cfg.resetOffset()),
		kgo.FetchMinBytes(int32(cfg.ConsumerMinBytes)),
		kgo.FetchMaxBytes(int32(cfg.ConsumerMaxBytes)),
		kgo.FetchMaxWait(cfg.ConsumerMaxWait),
		kgo.SessionTimeout(cfg.ConsumerSessionTimeout),
		kgo.HeartbeatInterval(cfg.ConsumerHeartbeatInterval),
		kgo.RebalanceTimeout(cfg.ConsumerRebalanceTimeout),
		kgo.DisableAutoCommit(),
	}
	if cfg.ConsumerRack != "" {
		opts = append(opts, kgo.Rack(cfg.ConsumerRack))
	}
	opts = append(opts, authOpts...)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	zap.L().Info("kafka consumer created",
		zap.String("topic", cfg.ConsumerTopic),
		zap.String("group", cfg.ConsumerGroupID),
		zap.String("auth", string(cfg.Auth)),
	)
	return &Consumer{client: client, chain: newChain(interceptors)}, nil
}

func (c *Consumer) Run(ctx context.Context, handler MessageHandler) {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			logger.FromContext(ctx).Error("kafka fetch error",
				zap.String("topic", topic),
				zap.Int32("partition", partition),
				zap.Error(err),
			)
		})

		fetches.EachRecord(func(rec *kgo.Record) {
			rctx := injectTIDFromHeaders(ctx, rec.Headers)

			if err := c.chain.Before(rctx, rec); err != nil {
				return
			}

			handlerErr := handler(rctx, rec)
			c.chain.After(rctx, rec, handlerErr)

			if handlerErr != nil {
				logger.FromContext(rctx).Error("kafka message handler failed", zap.Error(handlerErr))
				return
			}

			if err := c.client.CommitRecords(rctx, rec); err != nil {
				logger.FromContext(rctx).Error("kafka commit failed", zap.Error(err))
			}
		})
	}
}

func (c *Consumer) Close() error {
	c.client.Close()
	return nil
}

func injectTIDFromHeaders(ctx context.Context, headers []kgo.RecordHeader) context.Context {
	for _, h := range headers {
		if h.Key == logger.TIDKey && len(h.Value) > 0 {
			return logger.WithTID(ctx, string(h.Value))
		}
	}
	return ctx
}
