package config

import "github.com/caarlos0/env/v6"

type envConfig struct {
	EnvServerAddr      string `env:"SERVER_ADDRESS"`
	EnvBaseURL         string `env:"BASE_URL"`
	EnvLogLevel        string `env:"LOG_LEVEL"`
	EnvFileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func parseEnv() (*envConfig, error) {
	var cfg envConfig
	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, err
	}

	return &cfg, nil
}
