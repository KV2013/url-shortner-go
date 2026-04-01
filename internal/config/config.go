package config

import "github.com/caarlos0/env/v6"

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() (*Config, error) {

	cfg := Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "./urls.json",
	}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	parseFlags()

	if flagServerAddr != "" {
		cfg.ServerAddress = flagServerAddr
	}
	if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}
	if flagLogLevel != "" {
		cfg.LogLevel = flagLogLevel
	}
	if flagFileStoragePath != "" {
		cfg.FileStoragePath = flagFileStoragePath
	}

	return &cfg, nil
}
