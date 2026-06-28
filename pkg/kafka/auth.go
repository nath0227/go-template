package kafka

import (
	"context"
	"crypto/tls"
	"fmt"

	awsmsk "github.com/aws/aws-msk-iam-sasl-signer-go/signer"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

// buildAuthOpts returns kgo options for the configured authentication method.
func buildAuthOpts(cfg KafkaConfig) ([]kgo.Opt, error) {
	switch cfg.Auth {
	case AuthSASLScram512:
		mechanism := scram.Auth{User: cfg.Username, Pass: cfg.Password}.AsSha512Mechanism()
		return []kgo.Opt{
			kgo.SASL(mechanism),
			kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
		}, nil
	case AuthAWSMSK:
		return []kgo.Opt{
			kgo.SASL(&mskMechanism{region: cfg.AWSRegion}),
			kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
		}, nil
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

func (m *mskMechanism) Authenticate(ctx context.Context, _ string) (sasl.Session, []byte, error) {
	token, _, err := awsmsk.GenerateAuthToken(ctx, m.region)
	if err != nil {
		return nil, nil, fmt.Errorf("msk iam token: %w", err)
	}
	// RFC 7628 OAUTHBEARER initial client response
	payload := fmt.Sprintf("n,,\x01auth=Bearer %s\x01\x01", token)
	return mskSession{}, []byte(payload), nil
}

type mskSession struct{}

func (mskSession) Challenge(_ []byte) (bool, []byte, error) { return true, nil, nil }
