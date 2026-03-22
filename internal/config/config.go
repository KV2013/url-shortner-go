package config

type Config struct {
	ServerAddress string
	BaseURL       string
	LogLevel      string
}

func NewConfig() (*Config, error) {
	parseFlags()

	if flagBaseURL == "" {
		flagBaseURL = "http://" + flagRunAddr
	}

	cfg := Config{
		ServerAddress: flagRunAddr,
		BaseURL:       flagBaseURL,
		LogLevel:      flagLogLevel,
	}

	envCfg, err := parseEnv()
	if err != nil {
		return nil, err
	}
	if envCfg.EnvRunAddr != "" {
		cfg.ServerAddress = envCfg.EnvRunAddr
	}
	if envCfg.EnvBaseURL != "" {
		cfg.BaseURL = envCfg.EnvBaseURL
	}

	return &cfg, nil
}
