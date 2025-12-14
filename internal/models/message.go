package models

import "time"

type Message struct {
	ID        string    `json:"id"`
	Room      string    `json:"room"`
	User      string    `json:"user"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type SendMessageRequest struct {
	User    string `json:"user" validate:"required"`
	Content string `json:"content" validate:"required"`
}
