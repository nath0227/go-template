package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/your-org/service-name/pkg/logger"
)

type Producer struct {
	client *kgo.Client
	async  bool
	chain  *interceptorChain
}

func NewProducer(cfg KafkaConfig, interceptors ...Interceptor) (*Producer, error) {
	authOpts, err := buildAuthOpts(cfg)
	if err != nil {
		return nil, err
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(cfg.ProducerTopic),
		kgo.RecordPartitioner(cfg.partitioner()),
		kgo.RequiredAcks(cfg.acks()),
		kgo.ProducerBatchCompression(cfg.compressionCodec()),
		kgo.ProducerBatchMaxBytes(int32(cfg.ProducerBatchBytes)),
		kgo.ProducerLinger(cfg.ProducerBatchTimeout),
		kgo.ProduceRequestTimeout(cfg.ProducerWriteTimeout),
		kgo.RecordRetries(cfg.ProducerMaxAttempts),
	}
	opts = append(opts, authOpts...)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	zap.L().Info("kafka producer created",
		zap.String("topic", cfg.ProducerTopic),
		zap.String("auth", string(cfg.Auth)),
	)
	return &Producer{client: client, async: cfg.ProducerAsync, chain: newChain(interceptors)}, nil
}

func (p *Producer) Produce(ctx context.Context, recs ...*kgo.Record) error {
	log := logger.FromContext(ctx)
	for _, rec := range recs {
		if err := p.chain.Before(ctx, rec); err != nil {
			return err
		}
	}

	if p.async {
		for _, rec := range recs {
			r := rec
			p.client.Produce(ctx, r, func(rec *kgo.Record, err error) {
				if err != nil {
					log.Error("kafka produce failed", zap.Error(err))
				}
				p.chain.After(ctx, rec, err)
			})
		}
		return nil
	}

	results := p.client.ProduceSync(ctx, recs...)
	var firstErr error
	for _, res := range results {
		if res.Err != nil {
			log.Error("kafka produce failed", zap.Error(res.Err))
			if firstErr == nil {
				firstErr = res.Err
			}
		}
		p.chain.After(ctx, res.Record, res.Err)
	}
	return firstErr
}

func (p *Producer) Close() error {
	p.client.Close()
	return nil
}
