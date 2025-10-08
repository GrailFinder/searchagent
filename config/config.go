package config

import (
	"log/slog"

	"github.com/BurntSushi/toml"
)

type Config struct {
	SEARXAPI string `toml:"SEARX_API"`
	ServerPort int    `toml:"SERVER_PORT"`
}

func LoadConfigOrDefault(fn string) *Config {
	if fn == "" {
		fn = "config.toml"
	}
	config := &Config{}
	_, err := toml.DecodeFile(fn, &config)
	if err != nil {
		slog.Warn("failed to read config from file, loading default", "error", err)
		panic(err)
	}
	return config
}
