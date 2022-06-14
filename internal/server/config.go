package server

import (
	"l0/internal/store"
)

type Config struct {
	Store   *store.Config
	NatsURL string
}

func NewConfig(natsURL string, storeConfig *store.Config) *Config {
	return &Config{
		NatsURL: natsURL,
		Store:   storeConfig,
	}
}
