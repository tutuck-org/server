package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Port       int  `toml:"port"`
	MaxClients int  `toml:"max_clients"`
	LogDMs     bool `toml:"log_dms"`
}

var cfg Config

func loadConfig(path string) Config {
	cfg := Config{
		Port:       4756,
		MaxClients: 100,
		LogDMs:     false,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	_, err = toml.Decode(string(data), &cfg)
	if err != nil {
		return cfg
	}

	return cfg

}
