package solana

import (
	"context"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/kirill-a-belov/solana_token_manager/pkg/logger"
	"github.com/kirill-a-belov/solana_token_manager/pkg/tracer"
)

type config struct {
	UseSandbox bool `envconfig:"SOLANA_USE_SANDBOX"`

	ApiSandboxUrl string `envconfig:"SOLANA_API_SANDBOX_URL" default:"https://api.devnet.solana.com"`
	ApiUrl        string `envconfig:"SOLANA_API_URL" default:"https://api.mainnet-beta.solana.com"`
}

func (c *config) Load() error {
	return envconfig.Process("", c)
}

var m *Module

func New(ctx context.Context) (*Module, error) {
	_, span := tracer.Start(ctx, "internal.solana.New")
	defer span.End()

	if m != nil {
		return m, nil
	}

	l := logger.New("solana")

	c := &config{}
	if err := c.Load(); err != nil {
		return nil, errors.Wrap(err, "loading configuration")
	}

	apiUrl := c.ApiUrl
	if c.UseSandbox {
		apiUrl = c.ApiSandboxUrl
	}

	sc := client.NewClient(apiUrl)

	m = &Module{
		config: c,
		log:    l,

		solanaClient: sc,
	}

	return m, nil
}

type Module struct {
	config *config
	log    logger.Logger

	solanaClient *client.Client
}
