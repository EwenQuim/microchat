package handlers

import (
	"fmt"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/EwenQuim/microchat/pkg/crypto"

	"github.com/go-fuego/fuego"
)

func GetMessages(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.Message, error) {
	return func(c fuego.ContextNoBody) ([]models.Message, error) {
		room := c.PathParam("room")
		return chatService.GetMessages(c.Context(), room)
	}
}

func SendMessage(chatService *services.ChatService) func(c fuego.ContextWithBody[models.SendMessageRequest]) (*models.Message, error) {
	return func(c fuego.ContextWithBody[models.SendMessageRequest]) (*models.Message, error) {
		room := c.PathParam("room")
		body, err := c.Body()
		if err != nil {
			return nil, err
		}

		// Validate cryptographic signature if provided
		if body.Signature != "" && body.Pubkey != "" && body.Timestamp != 0 {
			err := crypto.VerifyMessageSignature(
				body.Pubkey,
				body.Signature,
				body.Content,
				room,
				body.Timestamp,
			)
			if err != nil {
				return nil, fmt.Errorf("signature verification failed: %w", err)
			}
		}

		return chatService.SendMessage(c.Context(), room, body.User, body.Content, body.Signature, body.Pubkey, body.Timestamp)
	}
}
