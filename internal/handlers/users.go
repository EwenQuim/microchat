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

		return chatService.RegisterUser(body.PublicKey)
	}
}

func GetUser(chatService *services.ChatService) func(c fuego.ContextNoBody) (*models.User, error) {
	return func(c fuego.ContextNoBody) (*models.User, error) {
		publicKey := c.PathParam("publicKey")
		return chatService.GetUser(publicKey)
	}
}

func GetAllUsers(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.User, error) {
	return func(c fuego.ContextNoBody) ([]models.User, error) {
		return chatService.GetAllUsers()
	}
}

func VerifyUser(chatService *services.ChatService) func(c fuego.ContextWithBody[models.VerifyUserRequest]) (string, error) {
	return func(c fuego.ContextWithBody[models.VerifyUserRequest]) (string, error) {
		body, err := c.Body()
		if err != nil {
			return "", err
		}

		err = chatService.VerifyUser(body.PublicKey)
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

		err = chatService.UnverifyUser(body.PublicKey)
		if err != nil {
			return "", err
		}

		return "User verification removed successfully", nil
	}
}
