package models

import "time"

type User struct {
	PublicKey string    `json:"public_key"` // Hex-encoded secp256k1 public key
	Verified  bool      `json:"verified"`   // Whether the public key has been verified
	CreatedAt time.Time `json:"created_at"` // When the user was registered
	UpdatedAt time.Time `json:"updated_at"` // When the user or key was last updated
}

type RegisterUserRequest struct {
	PublicKey string `json:"public_key" validate:"required"`
}

type VerifyUserRequest struct {
	PublicKey string `json:"public_key" validate:"required"`
}
