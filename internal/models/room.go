package models

type Room struct {
	Name                 string  `json:"name"`
	MessageCount         int     `json:"message_count"`
	Hidden               bool    `json:"hidden"`
	HasPassword          bool    `json:"has_password"`
	LastMessageContent   *string `json:"last_message_content,omitempty"`
	LastMessageUser      *string `json:"last_message_user,omitempty"`
	LastMessageTimestamp *string `json:"last_message_timestamp,omitempty"`
}

type CreateRoomRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=50"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=4,max=72"`
}

type UpdateRoomVisibilityRequest struct {
	Hidden    bool   `json:"hidden"`
	Pubkey    string `json:"pubkey" validate:"required"`
	Signature string `json:"signature" validate:"required"`
	Timestamp int64  `json:"timestamp" validate:"required"`
}
