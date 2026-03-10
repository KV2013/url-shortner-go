package config

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() (*Config, error) {
	parseFlags()

	if flagBaseURL == "" {
		flagBaseURL = flagRunAddr
	}

	return &Config{
		ServerAddress: flagRunAddr,
		BaseURL:       flagBaseURL,
	}, nil
}
