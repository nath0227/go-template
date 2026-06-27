package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/your-org/service-name/pkg/logger"
)

type Producer struct {
	writer *kafka.Writer
	chain  *interceptorChain
}

func NewProducer(cfg KafkaConfig, interceptors ...Interceptor) (*Producer, error) {
	transport, err := buildTransport(cfg)
	if err != nil {
		return nil, err
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.ProducerTopic,
		BatchSize:    cfg.ProducerBatchSize,
		BatchTimeout: cfg.ProducerBatchTimeout,
		Balancer:     &kafka.LeastBytes{},
	}
	if transport != nil {
		w.Transport = transport
	}

	zap.L().Info("kafka producer created",
		zap.String("topic", cfg.ProducerTopic),
		zap.String("auth", string(cfg.Auth)),
	)
	return &Producer{writer: w, chain: newChain(interceptors)}, nil
}

func (p *Producer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	log := logger.FromContext(ctx)
	for i := range msgs {
		if err := p.chain.Before(ctx, &msgs[i]); err != nil {
			return err
		}
	}
	err := p.writer.WriteMessages(ctx, msgs...)
	if err != nil {
		log.Error("kafka write messages failed", zap.Error(err))
	}
	for i := range msgs {
		p.chain.After(ctx, &msgs[i], err)
	}
	return err
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
