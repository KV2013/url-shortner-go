package config

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() *Config {
	parseFlags()

	if flagBaseURL == "" {
		flagBaseURL = "http://" + flagRunAddr
	}

	return &Config{
		ServerAddress: flagRunAddr,
		BaseURL:       flagBaseURL,
	}
}
