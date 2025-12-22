package models

type Room struct {
	Name                 string  `json:"name"`
	HasPassword          bool    `json:"has_password"`
	LastMessageContent   *string `json:"last_message_content,omitempty"`
	LastMessageUser      *string `json:"last_message_user,omitempty"`
	LastMessageTimestamp *string `json:"last_message_timestamp,omitempty"`
}

type CreateRoomRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=50"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=4,max=72"`
}
