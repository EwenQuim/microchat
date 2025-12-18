package models

type Room struct {
	Name         string `json:"name"`
	MessageCount int    `json:"message_count"`
	Hidden       bool   `json:"hidden"`
}

type CreateRoomRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

type UpdateRoomVisibilityRequest struct {
	Hidden    bool   `json:"hidden"`
	Pubkey    string `json:"pubkey" validate:"required"`
	Signature string `json:"signature" validate:"required"`
	Timestamp int64  `json:"timestamp" validate:"required"`
}
