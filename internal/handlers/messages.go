package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/EwenQuim/microchat/internal/middleware"
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/EwenQuim/microchat/pkg/crypto"

	"github.com/go-fuego/fuego"
)

type GetMessagesQuery struct {
	Password string `query:"password"`
	Limit    int    `query:"limit"`
	Before   string `query:"before"` // RFC3339
}

func GetMessages(chatService *services.ChatService, pwLimiter *middleware.RateLimiter) func(c fuego.ContextWithParams[GetMessagesQuery]) ([]models.Message, error) {
	return func(c fuego.ContextWithParams[GetMessagesQuery]) ([]models.Message, error) {
		room := c.PathParam("room")
		queryParams, err := c.Params() //nolint:staticcheck // no replacement available yet in fuego
		if err != nil {
			return nil, err
		}
		password := queryParams.Password

		err = chatService.ValidateRoomPassword(c.Context(), room, password)
		if err != nil {
			ip := middleware.IPFromRequest(c.Request())
			if !pwLimiter.Allow("pw:"+ip, 5, time.Minute) {
				return nil, fuego.HTTPError{Status: http.StatusTooManyRequests, Title: "Too Many Requests", Detail: "too many failed password attempts"}
			}
			slog.ErrorContext(c, "cannot validate password", "err", err)
			time.Sleep(500 * time.Millisecond) // Mitigate brute-force attacks
			return []models.Message{}, nil
		}

		msgParams := services.MessageQueryParams{
			Limit: queryParams.Limit,
		}
		if queryParams.Before != "" {
			t, parseErr := time.Parse(time.RFC3339, queryParams.Before)
			if parseErr != nil {
				return nil, fuego.HTTPError{Status: http.StatusBadRequest, Title: "Bad Request", Detail: "invalid 'before' timestamp: use RFC3339 format"}
			}
			msgParams.Before = &t
		}
		if msgParams.Limit > 200 {
			msgParams.Limit = 200
		}

		return chatService.GetMessages(c.Context(), room, msgParams)
	}
}

func SendMessage(chatService *services.ChatService, pwLimiter *middleware.RateLimiter) func(c fuego.ContextWithBody[models.SendMessageRequest]) (*models.Message, error) {
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
				ip := middleware.IPFromRequest(c.Request())
				if !pwLimiter.Allow("pw:"+ip, 5, time.Minute) {
					return nil, fuego.HTTPError{Status: http.StatusTooManyRequests, Title: "Too Many Requests", Detail: "too many failed password attempts"}
				}
				return nil, fmt.Errorf("invalid room password")
			}
		} else {
			// Check if room requires password
			err := chatService.ValidateRoomPassword(c.Context(), room, "")
			if err != nil && err.Error() == "invalid password" {
				return nil, fmt.Errorf("password required for this room")
			}
		}

		// Always verify — fuego validates required fields before we get here
		if err := crypto.VerifyMessageSignature(body.Pubkey, body.Signature, body.Content, room, body.Timestamp); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}

		return chatService.SendMessage(c.Context(), room, body.User, body.Content, body.Signature, body.Pubkey, body.Timestamp)
	}
}
