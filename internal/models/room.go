package models

type Room struct {
	Name         string `json:"name"`
	MessageCount int    `json:"message_count"`
}
