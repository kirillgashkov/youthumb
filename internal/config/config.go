package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

const (
	ModeDevelopment = "development"
	ModeProduction  = "production"
)

type Config struct {
	Mode string `env:"APP_MODE" envDefault:"development"`
	GRPC GRPCConfig
}

type GRPCConfig struct {
	Host string `env:"APP_GRPC_HOST" envDefault:"127.0.0.1"`
	Port int    `env:"APP_GRPC_PORT" envDefault:"50051"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Mode != ModeDevelopment && cfg.Mode != ModeProduction {
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
	return nil
}
