package models

import (
	"time"
)

type Message struct {
	ID              string    `json:"id"`
	Room            string    `json:"room"`
	User            string    `json:"user"`
	Content         string    `json:"content"`
	Timestamp       time.Time `json:"timestamp"`
	Signature       string    `json:"signature,omitempty"`        // Cryptographic signature (hex-encoded)
	Pubkey          string    `json:"pubkey,omitempty"`           // Public key used for signing (hex-encoded)
	SignedTimestamp int64     `json:"signed_timestamp,omitempty"` // Unix timestamp that was signed
}

type SendMessageRequest struct {
	User         string `json:"user" validate:"required"`
	Content      string `json:"content" validate:"required"`
	Signature    string `json:"signature" validate:"required"`
	Pubkey       string `json:"pubkey" validate:"required"`
	Timestamp    int64  `json:"timestamp" validate:"required"`
	RoomPassword string `json:"room_password,omitempty"`
}
