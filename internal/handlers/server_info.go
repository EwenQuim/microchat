package handlers

import (
	"github.com/EwenQuim/microchat/internal/config"
	"github.com/go-fuego/fuego"
)

type ServerInfoResponse struct {
	SuggestedQuickname string   `json:"suggested_quickname"`
	Description        string   `json:"description"`
	SuggestedServers   []string `json:"suggested_servers,omitempty"`
}

func GetServerInfo(cfg *config.Config) func(ctx fuego.ContextNoBody) (ServerInfoResponse, error) {
	return func(ctx fuego.ContextNoBody) (ServerInfoResponse, error) {
		return ServerInfoResponse{
			SuggestedQuickname: cfg.QuickName,
			Description:        cfg.Description,
			SuggestedServers:   cfg.SuggestedServerList,
		}, nil
	}
}
