package models

import (
	"context"

	"github.com/go-fuego/fuego"
)

type Room struct {
	Name                 string  `json:"name"`
	HasPassword          bool    `json:"has_password"`
	IsEncrypted          bool    `json:"is_encrypted"`
	EncryptionSalt       *string `json:"encryption_salt,omitempty"`
	LastMessageContent   *string `json:"last_message_content,omitempty"`
	LastMessageUser      *string `json:"last_message_user,omitempty"`
	LastMessageTimestamp *string `json:"last_message_timestamp,omitempty"`
}

var _ fuego.OutTransformer = (*Room)(nil)

// InTransform implements fuego.InTransformer.
func (r *Room) OutTransform(context.Context) error {
	if r.HasPassword {
		r.LastMessageContent = nil
		r.LastMessageUser = nil
		r.LastMessageTimestamp = nil
	}

	return nil
}

type CreateRoomRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=50"`
	Password    *string `json:"password,omitempty" validate:"omitempty,min=4,max=72"`
	IsEncrypted bool    `json:"is_encrypted"`
}
