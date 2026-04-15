package config

import "github.com/caarlos0/env/v6"

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DbHost          string `env:"DB_HOST"`
	DbPort          int    `env:"DB_PORT"`
	DbUser          string `env:"DB_USER"`
	DbPassword      string `env:"DB_PASSWORD"`
	DbName          string `env:"DB_NAME"`
	DbSSLMode       string `env:"DB_SSL_MODE"`
}

func NewConfig() (*Config, error) {

	cfg := Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "./urls.json",
		DbHost:          "localhost",
		DbPort:          5432,
		DbUser:          "user1",
		DbPassword:      "password1",
		DbName:          "urlshortenergo",
		DbSSLMode:       "disable",
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
	if flagDbHost != "" {
		cfg.DbHost = flagDbHost
	}
	if flagDbPort != 0 {
		cfg.DbPort = flagDbPort
	}
	if flagDbUser != "" {
		cfg.DbUser = flagDbUser
	}
	if flagDbPassword != "" {
		cfg.DbPassword = flagDbPassword
	}
	if flagDbName != "" {
		cfg.DbName = flagDbName
	}
	if flagDbSSLMode != "" {
		cfg.DbSSLMode = flagDbSSLMode
	}

	return &cfg, nil
}
