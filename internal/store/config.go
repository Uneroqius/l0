package store

type Config struct {
	DatabaseURL string
}

func NewConfig(databaseURL string) *Config {
	return &Config{
		DatabaseURL: databaseURL,
	}
}
