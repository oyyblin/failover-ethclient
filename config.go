package ethclient

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

const (
	DefaultEnvPrefix = "ethclient"
	nameLenLimit     = 32
)

type Config struct {
	EnablePrometheus bool   `default:"true"`
	RpcUrl           string `required:"true"`
	RpcName          string `required:"true"`
	FailoverRpcUrl   string `required:"true"`
	FailoverRpcName  string `required:"true"`
}

func (c *Config) Valid() error {
	if len(c.RpcName) >= nameLenLimit {
		return fmt.Errorf("invalid RpcName: %s", c.RpcName)
	}
	if len(c.FailoverRpcName) >= nameLenLimit {
		return fmt.Errorf("invalid RpcFailoverName: %s", c.FailoverRpcName)
	}
	return nil
}

func ConfigFromEnv() *Config {
	config := &Config{}
	envconfig.MustProcess(DefaultEnvPrefix, config)
	if err := config.Valid(); err != nil {
		log.Fatal().Msgf("%s", err)
	}
	return config
}

func ConfigFromEnvPrefix(prefix string) *Config {
	config := &Config{}
	envconfig.MustProcess(prefix, config)
	if err := config.Valid(); err != nil {
		log.Fatal().Msgf("%s", err)
	}
	return config
}
