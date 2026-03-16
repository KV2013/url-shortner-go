package config

import "github.com/caarlos0/env/v6"

type envConfig struct {
	EnvRunAddr string `env:"SERVER_ADDRESS"`
	EnvBaseURL string `env:"BASE_URL"`
}

func parseEnv() (*envConfig, error) {
	var cfg envConfig
	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, err
	}

	return &cfg, nil
}
