package handlers

import (
	"context"
	"strings"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

// filterVisibleRooms hides password-protected rooms unless they appear in the
// comma-separated visitedParam, and applies OutTransform to each visible room.
func filterVisibleRooms(ctx context.Context, allRooms []models.Room, visitedParam string) ([]models.Room, error) {
	visitedRooms := make(map[string]bool)
	if visitedParam != "" {
		for roomName := range strings.SplitSeq(visitedParam, ",") {
			if roomName != "" {
				visitedRooms[roomName] = true
			}
		}
	}

	visibleRooms := make([]models.Room, 0)
	for _, room := range allRooms {
		// Hide password-protected rooms unless already visited
		if room.HasPassword && !visitedRooms[room.Name] {
			continue
		}
		if err := room.OutTransform(ctx); err != nil {
			return nil, err
		}
		visibleRooms = append(visibleRooms, room)
	}

	return visibleRooms, nil
}

type GetRoomsQuery struct {
	Visited string `query:"visited"`
}

func GetRooms(chatService *services.ChatService) func(c fuego.ContextWithParams[GetRoomsQuery]) ([]models.Room, error) {
	return func(c fuego.ContextWithParams[GetRoomsQuery]) ([]models.Room, error) {
		allRooms, err := chatService.GetRooms(c.Context())
		if err != nil {
			return nil, err
		}

		params, err := c.Params() //nolint:staticcheck // no replacement available yet in fuego
		if err != nil {
			return nil, err
		}

		return filterVisibleRooms(c.Context(), allRooms, params.Visited)
	}
}

type SearchRoomsQuery struct {
	Visited string `query:"visited"`
	Q       string `query:"q"`
}

func SearchRooms(chatService *services.ChatService) func(c fuego.ContextWithParams[SearchRoomsQuery]) ([]models.Room, error) {
	return func(c fuego.ContextWithParams[SearchRoomsQuery]) ([]models.Room, error) {
		params, err := c.Params() //nolint:staticcheck // no replacement available yet in fuego
		if err != nil {
			return nil, err
		}

		allRooms, err := chatService.SearchRooms(c.Context(), params.Q)
		if err != nil {
			return nil, err
		}

		return filterVisibleRooms(c.Context(), allRooms, params.Visited)
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
