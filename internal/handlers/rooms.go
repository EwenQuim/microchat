package handlers

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/EwenQuim/microchat/pkg/crypto"

	"github.com/go-fuego/fuego"
)

type GetRoomsQuery struct {
	Visited string `query:"visited"`
}

func GetRooms(chatService *services.ChatService) func(c fuego.ContextWithParams[GetRoomsQuery]) ([]models.Room, error) {
	return func(c fuego.ContextWithParams[GetRoomsQuery]) ([]models.Room, error) {
		allRooms, err := chatService.GetRooms(c.Context())
		if err != nil {
			return nil, err
		}

		params, err := c.Params()
		if err != nil {
			return nil, err
		}
		// Get list of visited rooms from query parameter (comma-separated)
		visitedParam := params.Visited
		visitedRooms := make(map[string]bool)
		if visitedParam != "" {
			for roomName := range strings.SplitSeq(visitedParam, ",") {
				if roomName != "" {
					visitedRooms[roomName] = true
				}
			}
		}

		// Filter out hidden rooms and password-protected rooms (unless visited)
		visibleRooms := make([]models.Room, 0)
		for _, room := range allRooms {
			if room.Hidden {
				continue
			}
			// Hide password-protected rooms unless already visited
			if room.HasPassword && !visitedRooms[room.Name] {
				continue
			}
			visibleRooms = append(visibleRooms, room)
		}

		return visibleRooms, nil
	}
}

type SearchRoomsQuery struct {
	Visited string `query:"visited"`
	Q       string `query:"q"`
}

func SearchRooms(chatService *services.ChatService) func(c fuego.ContextWithParams[SearchRoomsQuery]) ([]models.Room, error) {
	return func(c fuego.ContextWithParams[SearchRoomsQuery]) ([]models.Room, error) {
		params, err := c.Params()
		if err != nil {
			return nil, err
		}

		allRooms, err := chatService.SearchRooms(c.Context(), params.Q)
		if err != nil {
			return nil, err
		}

		// Get list of visited rooms from query parameter (comma-separated)
		visitedParam := params.Visited
		visitedRooms := make(map[string]bool)
		if visitedParam != "" {
			for roomName := range strings.SplitSeq(visitedParam, ",") {
				if roomName != "" {
					visitedRooms[roomName] = true
				}
			}
		}

		// Filter out hidden rooms and password-protected rooms (unless visited)
		visibleRooms := make([]models.Room, 0)
		for _, room := range allRooms {
			if room.Hidden {
				continue
			}
			// Hide password-protected rooms unless already visited
			if room.HasPassword && !visitedRooms[room.Name] {
				continue
			}
			visibleRooms = append(visibleRooms, room)
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
		return chatService.CreateRoom(c.Context(), body.Name, body.Password)
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
