package handlers

import (
	"strings"

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

		// Filter out password-protected rooms (unless visited)
		visibleRooms := make([]models.Room, 0)
		for _, room := range allRooms {
			// Hide password-protected rooms unless already visited
			if room.HasPassword && !visitedRooms[room.Name] {
				continue
			}
			err = room.OutTransform(c)
			if err != nil {
				return nil, err
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

		// Filter out password-protected rooms (unless visited)
		visibleRooms := make([]models.Room, 0)
		for _, room := range allRooms {
			// Hide password-protected rooms unless already visited
			if room.HasPassword && !visitedRooms[room.Name] {
				continue
			}
			err = room.OutTransform(c)
			if err != nil {
				return nil, err
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

		// Generate encryption salt if encryption is enabled
		var encryptionSalt *string
		if body.IsEncrypted {
			salt := crypto.GenerateRandomHex(32) // 32 bytes for salt
			encryptionSalt = &salt
		}

		return chatService.CreateRoom(c.Context(), body.Name, body.Password, body.IsEncrypted, encryptionSalt)
	}
}
