package config

import "os"

type Config struct {
	Token string
}

func GetConfig() *Config {
	var cfg Config

	cfg.Token = os.Getenv("TOKEN")

	return &cfg
}
