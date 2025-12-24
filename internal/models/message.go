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
	IsEncrypted     bool      `json:"is_encrypted"`
	Nonce           string    `json:"nonce,omitempty"` // Hex-encoded nonce for AES-GCM
}

type SendMessageRequest struct {
	User         string `json:"user" validate:"required"`
	Content      string `json:"content" validate:"required"`
	Signature    string `json:"signature,omitempty"`
	Pubkey       string `json:"pubkey,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
	RoomPassword string `json:"room_password,omitempty"`
	IsEncrypted  bool   `json:"is_encrypted"`
	Nonce        string `json:"nonce,omitempty"`
}
