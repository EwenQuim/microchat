package handlers

import (
	"fmt"
	"slices"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/EwenQuim/microchat/pkg/crypto"

	"github.com/go-fuego/fuego"
)

func GetRooms(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.Room, error) {
	return func(c fuego.ContextNoBody) ([]models.Room, error) {
		allRooms, err := chatService.GetRooms(c.Context())
		if err != nil {
			return nil, err
		}

		// Filter out hidden rooms
		visibleRooms := make([]models.Room, 0)
		for _, room := range allRooms {
			if !room.Hidden {
				visibleRooms = append(visibleRooms, room)
			}
		}

		return visibleRooms, nil
	}
}

func SearchRooms(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.Room, error) {
	return func(c fuego.ContextNoBody) ([]models.Room, error) {
		query := c.QueryParam("q")
		allRooms, err := chatService.SearchRooms(c.Context(), query)
		if err != nil {
			return nil, err
		}

		// Filter out hidden rooms
		visibleRooms := make([]models.Room, 0)
		for _, room := range allRooms {
			if !room.Hidden {
				visibleRooms = append(visibleRooms, room)
			}
		}

		return visibleRooms, nil
	}
}

func CreateRoom(chatService *services.ChatService) func(c fuego.ContextWithBody[models.CreateRoomRequest]) (*models.Room, error) {
	return func(c fuego.ContextWithBody[models.CreateRoomRequest]) (*models.Room, error) {
		body, err := c.Body()
		if err != nil {
			return nil, err
		}
		return chatService.CreateRoom(c.Context(), body.Name)
	}
}

func UpdateRoomVisibility(chatService *services.ChatService, cfg *config.Config) func(c fuego.ContextWithBody[models.UpdateRoomVisibilityRequest]) (string, error) {
	return func(c fuego.ContextWithBody[models.UpdateRoomVisibilityRequest]) (string, error) {
		room := c.PathParam("room")
		body, err := c.Body()
		if err != nil {
			return "", err
		}

		// Verify that all required fields are present
		if body.Pubkey == "" || body.Signature == "" || body.Timestamp == 0 {
			return "", fmt.Errorf("pubkey, signature, and timestamp are required")
		}

		// Verify signature
		err = crypto.VerifyRoomVisibilitySignature(
			body.Pubkey,
			body.Signature,
			room,
			body.Hidden,
			body.Timestamp,
		)
		if err != nil {
			return "", fmt.Errorf("signature verification failed: %w", err)
		}

		// Check if pubkey is in admin list
		isAdmin := slices.Contains(cfg.AdminPubkeys, body.Pubkey)

		if !isAdmin {
			return "", fmt.Errorf("unauthorized: only admins can update room visibility")
		}

		// Update room visibility
		err = chatService.UpdateRoomVisibility(c.Context(), room, body.Hidden)
		if err != nil {
			return "", err
		}

		return "Room visibility updated successfully", nil
	}
}
