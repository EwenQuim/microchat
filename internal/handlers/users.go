package handlers

import (
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func RegisterUser(chatService *services.ChatService) func(c fuego.ContextWithBody[models.RegisterUserRequest]) (*models.User, error) {
	return func(c fuego.ContextWithBody[models.RegisterUserRequest]) (*models.User, error) {
		body, err := c.Body()
		if err != nil {
			return nil, err
		}

		return chatService.RegisterUser(c.Context(), body.PublicKey)
	}
}

func GetUser(chatService *services.ChatService) func(c fuego.ContextNoBody) (*models.User, error) {
	return func(c fuego.ContextNoBody) (*models.User, error) {
		publicKey := c.PathParam("publicKey")
		return chatService.GetUser(c.Context(), publicKey)
	}
}

func GetAllUsers(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.User, error) {
	return func(c fuego.ContextNoBody) ([]models.User, error) {
		return chatService.GetAllUsers(c.Context())
	}
}

func VerifyUser(chatService *services.ChatService) func(c fuego.ContextWithBody[models.VerifyUserRequest]) (string, error) {
	return func(c fuego.ContextWithBody[models.VerifyUserRequest]) (string, error) {
		body, err := c.Body()
		if err != nil {
			return "", err
		}

		err = chatService.VerifyUser(c.Context(), body.PublicKey)
		if err != nil {
			return "", err
		}

		return "User verified successfully", nil
	}
}

func UnverifyUser(chatService *services.ChatService) func(c fuego.ContextWithBody[models.VerifyUserRequest]) (string, error) {
	return func(c fuego.ContextWithBody[models.VerifyUserRequest]) (string, error) {
		body, err := c.Body()
		if err != nil {
			return "", err
		}

		err = chatService.UnverifyUser(c.Context(), body.PublicKey)
		if err != nil {
			return "", err
		}

		return "User verification removed successfully", nil
	}
}
