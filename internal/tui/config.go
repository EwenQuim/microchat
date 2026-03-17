package tui

import (
	"encoding/json"
	"fmt"
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

type userEntry struct {
	PubKey      string `json:"pub_key"`
	DisplayName string `json:"display_name"`
}

type appConfig struct {
	Identity   *identityConfig `json:"identity,omitempty"`
	Servers    []serverConfig  `json:"servers"`
	LastServer string          `json:"last_server,omitempty"`
	Users      []userEntry     `json:"users,omitempty"`
}

func checkConfigPermissions() error {
	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	dirPath := filepath.Dir(path)
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	if dirInfo.Mode().Perm() != 0700 {
		return fmt.Errorf("config directory has insecure permissions %04o (want 0700)\n  fix: chmod 0700 ~/.config/microchat/", dirInfo.Mode().Perm())
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.Mode().Perm() != 0600 {
		return fmt.Errorf("config file has insecure permissions %04o (want 0600)\n  fix: chmod 0600 ~/.config/microchat/config.json", fileInfo.Mode().Perm())
	}
	return nil
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
