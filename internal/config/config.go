package config

import (
	"cmp"
	"os"
	"strings"
)

type Config struct {
	Port                string
	AdminPubkeys        []string
	QuickName           string
	Description         string
	SuggestedServerList []string
}

func Load() *Config {
	port := cmp.Or(os.Getenv("PORT"), ":8080")

	hostname, _ := os.Hostname()
	quickName := cmp.Or(os.Getenv("SERVER_QUICKNAME"), hostname)
	description := os.Getenv("SERVER_DESCRIPTION")

	// Parse comma-separated list of admin public keys
	adminPubkeysStr := os.Getenv("ADMIN_PUBKEYS")
	var adminPubkeys []string
	if adminPubkeysStr != "" {
		for pk := range strings.SplitSeq(adminPubkeysStr, ",") {
			trimmed := strings.TrimSpace(pk)
			if trimmed != "" {
				adminPubkeys = append(adminPubkeys, trimmed)
			}
		}
	}

	// Parse comma-separated list of suggested server URLs
	suggestedServerListStr := os.Getenv("SUGGESTED_SERVER_LIST")
	var suggestedServerList []string
	if suggestedServerListStr != "" {
		for url := range strings.SplitSeq(suggestedServerListStr, ",") {
			trimmed := strings.TrimSpace(url)
			if trimmed != "" {
				suggestedServerList = append(suggestedServerList, trimmed)
			}
		}
	}

	return &Config{
		Port:                port,
		AdminPubkeys:        adminPubkeys,
		QuickName:           quickName,
		Description:         description,
		SuggestedServerList: suggestedServerList,
	}
}
