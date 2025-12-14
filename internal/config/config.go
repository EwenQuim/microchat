package config

import (
	"cmp"
	"os"
)

type Config struct {
	Port string
}

func Load() *Config {
	port := cmp.Or(os.Getenv("PORT"), "8080")

	return &Config{
		Port: port,
	}
}
