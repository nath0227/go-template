package kafka

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// AuthType defines the Kafka broker authentication method.
type AuthType string

const (
	AuthNone         AuthType = "none"
	AuthSASLScram512 AuthType = "sasl_scram_512"
	AuthAWSMSK       AuthType = "aws_msk" // AWS MSK IAM via OAUTHBEARER
)

// BalancerType selects the producer partition assignment strategy.
type BalancerType string

const (
	BalancerLeastBytes BalancerType = "least_bytes"
	BalancerRoundRobin BalancerType = "round_robin"
	BalancerHash       BalancerType = "hash"    // consistent hash on Record.Key
	BalancerMurmur2    BalancerType = "murmur2" // Kafka-compatible murmur2 hash on Record.Key
)

// AcksType controls producer write durability.
type AcksType string

const (
	AcksNone AcksType = "none" // fire-and-forget, highest throughput
	AcksOne  AcksType = "one"  // leader ack (default)
	AcksAll  AcksType = "all"  // all in-sync replica acks, safest
)

// CompressionType selects the producer message compression codec.
type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionGzip   CompressionType = "gzip"
	CompressionSnappy CompressionType = "snappy"
	CompressionLz4    CompressionType = "lz4"
	CompressionZstd   CompressionType = "zstd"
)

// GroupBalancerType selects the consumer group rebalance strategy.
type GroupBalancerType string

const (
	GroupBalancerCooperativeSticky GroupBalancerType = "cooperative_sticky" // franz-go default; no stop-the-world rebalance
	GroupBalancerRange             GroupBalancerType = "range"
	GroupBalancerRoundRobin        GroupBalancerType = "round_robin"
	GroupBalancerRackAffinity      GroupBalancerType = "rack_affinity" // cooperative-sticky + kgo.Rack()
)

// StartOffsetType controls where a NEW consumer group begins reading when no committed offset exists.
type StartOffsetType string

const (
	StartOffsetFirst StartOffsetType = "first" // earliest available message
	StartOffsetLast  StartOffsetType = "last"  // skip history; only messages produced after join
)

// KafkaConfig holds all Kafka settings — shared between producer and consumer.
type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envSeparator:","`

	// Auth
	Auth      AuthType `env:"KAFKA_AUTH" envDefault:"none"`
	Username  string   `env:"KAFKA_USERNAME"`
	Password  string   `env:"KAFKA_PASSWORD"`
	AWSRegion string   `env:"KAFKA_AWS_REGION"` // required when Auth == AuthAWSMSK

	// Producer
	ProducerTopic        string          `env:"KAFKA_PRODUCER_TOPIC"`
	ProducerBalancer     BalancerType    `env:"KAFKA_PRODUCER_BALANCER"      envDefault:"least_bytes"`
	ProducerRequiredAcks AcksType        `env:"KAFKA_PRODUCER_REQUIRED_ACKS" envDefault:"one"`
	ProducerCompression  CompressionType `env:"KAFKA_PRODUCER_COMPRESSION"   envDefault:"none"`
	ProducerBatchSize    int             `env:"KAFKA_PRODUCER_BATCH_SIZE"    envDefault:"100"`     // unused by franz-go; batching is automatic
	ProducerBatchBytes   int64           `env:"KAFKA_PRODUCER_BATCH_BYTES"   envDefault:"1048576"` // 1 MB
	ProducerBatchTimeout time.Duration   `env:"KAFKA_PRODUCER_BATCH_TIMEOUT" envDefault:"1ms"`
	ProducerMaxAttempts  int             `env:"KAFKA_PRODUCER_MAX_ATTEMPTS"  envDefault:"10"`
	ProducerWriteTimeout time.Duration   `env:"KAFKA_PRODUCER_WRITE_TIMEOUT" envDefault:"10s"`
	ProducerAsync        bool            `env:"KAFKA_PRODUCER_ASYNC"         envDefault:"false"`

	// Consumer
	ConsumerTopic             string            `env:"KAFKA_CONSUMER_TOPIC"`
	ConsumerGroupID           string            `env:"KAFKA_CONSUMER_GROUP_ID"`
	ConsumerGroupBalancer     GroupBalancerType `env:"KAFKA_CONSUMER_GROUP_BALANCER"      envDefault:"cooperative_sticky"`
	ConsumerRack              string            `env:"KAFKA_CONSUMER_RACK"`               // AZ/rack ID, used with rack_affinity balancer
	ConsumerStartOffset       StartOffsetType   `env:"KAFKA_CONSUMER_START_OFFSET"        envDefault:"first"`
	ConsumerMinBytes          int               `env:"KAFKA_CONSUMER_MIN_BYTES"           envDefault:"1"`
	ConsumerMaxBytes          int               `env:"KAFKA_CONSUMER_MAX_BYTES"           envDefault:"10485760"` // 10 MB
	ConsumerMaxWait           time.Duration     `env:"KAFKA_CONSUMER_MAX_WAIT"            envDefault:"10s"`
	ConsumerCommitInterval    time.Duration     `env:"KAFKA_CONSUMER_COMMIT_INTERVAL"     envDefault:"1s"` // unused; commits are explicit per-record
	ConsumerSessionTimeout    time.Duration     `env:"KAFKA_CONSUMER_SESSION_TIMEOUT"     envDefault:"30s"`
	ConsumerHeartbeatInterval time.Duration     `env:"KAFKA_CONSUMER_HEARTBEAT_INTERVAL"  envDefault:"3s"`
	ConsumerRebalanceTimeout  time.Duration     `env:"KAFKA_CONSUMER_REBALANCE_TIMEOUT"   envDefault:"30s"`
	ConsumerReadBackoffMin    time.Duration     `env:"KAFKA_CONSUMER_READ_BACKOFF_MIN"    envDefault:"100ms"` // unused; franz-go handles internally
	ConsumerReadBackoffMax    time.Duration     `env:"KAFKA_CONSUMER_READ_BACKOFF_MAX"    envDefault:"1s"`    // unused; franz-go handles internally
	ConsumerMaxAttempts       int               `env:"KAFKA_CONSUMER_MAX_ATTEMPTS"        envDefault:"3"`     // unused; franz-go handles internally
}

func (c KafkaConfig) partitioner() kgo.Partitioner {
	switch c.ProducerBalancer {
	case BalancerRoundRobin:
		return kgo.RoundRobinPartitioner()
	case BalancerHash, BalancerMurmur2:
		return kgo.StickyKeyPartitioner(nil) // murmur2 on key by default
	default: // BalancerLeastBytes
		return kgo.LeastBackupPartitioner()
	}
}

func (c KafkaConfig) acks() kgo.Acks {
	switch c.ProducerRequiredAcks {
	case AcksNone:
		return kgo.NoAck()
	case AcksAll:
		return kgo.AllISRAcks()
	default: // AcksOne
		return kgo.LeaderAck()
	}
}

func (c KafkaConfig) compressionCodec() kgo.CompressionCodec {
	switch c.ProducerCompression {
	case CompressionGzip:
		return kgo.GzipCompression()
	case CompressionSnappy:
		return kgo.SnappyCompression()
	case CompressionLz4:
		return kgo.Lz4Compression()
	case CompressionZstd:
		return kgo.ZstdCompression()
	default:
		return kgo.NoCompression()
	}
}

func (c KafkaConfig) balancers() []kgo.GroupBalancer {
	switch c.ConsumerGroupBalancer {
	case GroupBalancerRange:
		return []kgo.GroupBalancer{kgo.RangeBalancer(), kgo.RoundRobinBalancer()}
	case GroupBalancerRoundRobin:
		return []kgo.GroupBalancer{kgo.RoundRobinBalancer()}
	case GroupBalancerRackAffinity:
		// rack preference is handled via kgo.Rack() in consumer opts
		return []kgo.GroupBalancer{kgo.CooperativeStickyBalancer()}
	default: // GroupBalancerCooperativeSticky and unknown values
		return []kgo.GroupBalancer{kgo.CooperativeStickyBalancer()}
	}
}

func (c KafkaConfig) resetOffset() kgo.Offset {
	if c.ConsumerStartOffset == StartOffsetLast {
		return kgo.NewOffset().AtEnd()
	}
	return kgo.NewOffset().AtStart()
}
