package config

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress      string `env:"SERVER_ADDRESS"      json:"server_address"`
	BaseURL            string `env:"BASE_URL"            json:"base_url"`
	LogLevel           string `env:"LOG_LEVEL"           json:"log_level"`
	FileStoragePath    string `env:"FILE_STORAGE_PATH"   json:"file_storage_path"`
	DatabaseDSN        string `env:"DATABASE_DSN"        json:"database_dsn"`
	JWTSecretKey       string `env:"JWT_SECRET_KEY"      json:"jwt_secret_key"`
	AuditFile          string `env:"AUDIT_FILE"          json:"audit_file"`
	AuditURL           string `env:"AUDIT_URL"           json:"audit_url"`
	EnablePprof        bool   `env:"ENABLE_PPROF"        json:"enable_pprof"`
	EnableHTTPS        bool   `env:"ENABLE_HTTPS"        json:"enable_https"`
	TrustedSubnet      string `env:"TRUSTED_SUBNET"      json:"trusted_subnet"`
	AuditMaxConcurrent int    `env:"AUDIT_MAX_CONCURRENT" json:"audit_max_concurrent"`
}

func NewConfig() (*Config, error) {

	parseFlags()

	cfg := Config{}

	if err := LoadConfigFile(&cfg); err != nil {
		return nil, err
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = "localhost:8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.JWTSecretKey == "" {
		cfg.JWTSecretKey = "default-secret-change-me"
	}
	if cfg.AuditMaxConcurrent == 0 {
		cfg.AuditMaxConcurrent = 10
	}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

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
	if flagEnableHTTPS {
		cfg.EnableHTTPS = true
	}
	if flagTrustedSubnet != "" {
		cfg.TrustedSubnet = flagTrustedSubnet
	}

	if cfg.TrustedSubnet != "" {
		if _, _, err := net.ParseCIDR(cfg.TrustedSubnet); err != nil {
			return nil, fmt.Errorf("некорректный trusted_subnet: %w", err)
		}
	}

	return &cfg, nil
}
