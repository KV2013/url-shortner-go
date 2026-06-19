package config

import "github.com/caarlos0/env/v6"

type Config struct {
	ServerAddress       string `env:"SERVER_ADDRESS"`
	BaseURL             string `env:"BASE_URL"`
	LogLevel            string `env:"LOG_LEVEL"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN         string `env:"DATABASE_DSN"`
	JWTSecretKey        string `env:"JWT_SECRET_KEY"`
	AuditFile           string `env:"AUDIT_FILE"`
	AuditURL            string `env:"AUDIT_URL"`
	EnablePprof         bool   `env:"ENABLE_PPROF"`
	AuditMaxConcurrent  int    `env:"AUDIT_MAX_CONCURRENT"`
}

func NewConfig() (*Config, error) {

	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		LogLevel:      "info",
		JWTSecretKey:       "default-secret-change-me",
		AuditMaxConcurrent: 10,
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
	if flagDatabaseDSN != "" {
		cfg.DatabaseDSN = flagDatabaseDSN
	}
	if flagJWTSecretKey != "" {
		cfg.JWTSecretKey = flagJWTSecretKey
	}
	if flagAuditFile != "" {
		cfg.AuditFile = flagAuditFile
	}
	if flagAuditURL != "" {
		cfg.AuditURL = flagAuditURL
	}
	if flagEnablePprof {
		cfg.EnablePprof = true
	}

	return &cfg, nil
}
