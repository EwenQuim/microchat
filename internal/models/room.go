package models

type Room struct {
	Name         string `json:"name"`
	MessageCount int    `json:"message_count"`
}

type CreateRoomRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}
