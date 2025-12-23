package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/EwenQuim/microchat/pkg/crypto"

	"github.com/go-fuego/fuego"
)

type GetMessagesQuery struct {
	Password string `query:"password"`
}

func GetMessages(chatService *services.ChatService) func(c fuego.ContextWithParams[GetMessagesQuery]) ([]models.Message, error) {
	return func(c fuego.ContextWithParams[GetMessagesQuery]) ([]models.Message, error) {
		room := c.PathParam("room")
		params, err := c.Params()
		if err != nil {
			return nil, err
		}
		password := params.Password

		err = chatService.ValidateRoomPassword(c.Context(), room, password)
		if err != nil {
			slog.ErrorContext(c, "cannot validate password", "err", err)
			time.Sleep(500 * time.Millisecond) // Mitigate brute-force attacks
			return []models.Message{}, nil
		}

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

		// Get password from header or query param
		password := body.RoomPassword

		// Validate room password if provided
		if password != "" {
			err := chatService.ValidateRoomPassword(c.Context(), room, password)
			if err != nil {
				return nil, fmt.Errorf("invalid room password")
			}
		} else {
			// Check if room requires password
			err := chatService.ValidateRoomPassword(c.Context(), room, "")
			if err != nil && err.Error() == "invalid password" {
				return nil, fmt.Errorf("password required for this room")
			}
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
