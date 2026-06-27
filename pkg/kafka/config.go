package kafka

import "time"

// AuthType defines the Kafka broker authentication method.
type AuthType string

const (
	AuthNone         AuthType = "none"
	AuthSASLScram512 AuthType = "sasl_scram_512"
	AuthAWSMSK       AuthType = "aws_msk" // AWS MSK IAM via OAUTHBEARER
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
	ProducerTopic        string        `env:"KAFKA_PRODUCER_TOPIC"`
	ProducerBatchSize    int           `env:"KAFKA_PRODUCER_BATCH_SIZE" envDefault:"100"`
	ProducerBatchTimeout time.Duration `env:"KAFKA_PRODUCER_BATCH_TIMEOUT" envDefault:"1ms"`

	// Consumer
	ConsumerTopic          string        `env:"KAFKA_CONSUMER_TOPIC"`
	ConsumerGroupID        string        `env:"KAFKA_CONSUMER_GROUP_ID"`
	ConsumerMinBytes       int           `env:"KAFKA_CONSUMER_MIN_BYTES" envDefault:"1"`
	ConsumerMaxBytes       int           `env:"KAFKA_CONSUMER_MAX_BYTES" envDefault:"10485760"`
	ConsumerCommitInterval time.Duration `env:"KAFKA_CONSUMER_COMMIT_INTERVAL" envDefault:"1s"`
}
