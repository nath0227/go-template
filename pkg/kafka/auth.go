package kafka

import (
	"context"
	"crypto/tls"
	"fmt"

	awsmsk "github.com/aws/aws-msk-iam-sasl-signer-go/signer"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// buildTransport returns a Transport configured for the given auth method.
// Returns nil for AuthNone so kafka-go uses its default transport.
func buildTransport(cfg KafkaConfig) (kafka.RoundTripper, error) {
	mechanism, err := buildMechanism(cfg)
	if err != nil {
		return nil, err
	}
	if mechanism == nil {
		return nil, nil
	}
	return &kafka.Transport{
		SASL: mechanism,
		TLS:  &tls.Config{MinVersion: tls.VersionTLS12},
	}, nil
}

// buildDialer returns a Dialer configured for the given auth method.
// Returns nil for AuthNone so kafka-go uses its default dialer.
func buildDialer(cfg KafkaConfig) (*kafka.Dialer, error) {
	mechanism, err := buildMechanism(cfg)
	if err != nil {
		return nil, err
	}
	if mechanism == nil {
		return nil, nil
	}
	return &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{MinVersion: tls.VersionTLS12},
	}, nil
}

func buildMechanism(cfg KafkaConfig) (sasl.Mechanism, error) {
	switch cfg.Auth {
	case AuthSASLScram512:
		return scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
	case AuthAWSMSK:
		return &mskMechanism{region: cfg.AWSRegion}, nil
	default:
		return nil, nil
	}
}

// mskMechanism implements sasl.Mechanism for AWS MSK IAM (OAUTHBEARER).
// GenerateAuthToken is called on every new connection, providing implicit token refresh.
type mskMechanism struct {
	region string
}

func (m *mskMechanism) Name() string { return "OAUTHBEARER" }

func (m *mskMechanism) Start(ctx context.Context) (sasl.StateMachine, []byte, error) {
	token, _, err := awsmsk.GenerateAuthToken(ctx, m.region)
	if err != nil {
		return nil, nil, fmt.Errorf("msk iam token: %w", err)
	}
	// RFC 7628 OAUTHBEARER initial client response
	payload := fmt.Sprintf("n,,\x01auth=Bearer %s\x01\x01", token)
	return mskDone{}, []byte(payload), nil
}

type mskDone struct{}

func (mskDone) Next(_ context.Context, _ []byte) (bool, []byte, error) { return true, nil, nil }
