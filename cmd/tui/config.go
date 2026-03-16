package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type identityConfig struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type serverConfig struct {
	URL         string `json:"url"`
	Quickname   string `json:"quickname"`
	Description string `json:"description"`
}

type appConfig struct {
	Identity   *identityConfig `json:"identity,omitempty"`
	Servers    []serverConfig  `json:"servers"`
	LastServer string          `json:"last_server,omitempty"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "microchat", "config.json")
}

func loadConfig() (appConfig, error) {
	var cfg appConfig
	data, err := os.ReadFile(configPath())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	return cfg, json.Unmarshal(data, &cfg)
}

func saveConfig(cfg appConfig) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
