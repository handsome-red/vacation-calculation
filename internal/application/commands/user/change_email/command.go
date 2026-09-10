package change_email

import (
	"github.com/google/uuid"
	"time"
)

type Command struct {
	UserID   uuid.UUID
	NewEmail string
}

type Result struct {
	UserID    uuid.UUID `json:"user_id"`
	NewEmail  string    `json:"new_email"`
	UpdatedAt time.Time `json:"updated_at"`
}
