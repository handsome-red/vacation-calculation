package activate_user

import "github.com/google/uuid"

type Command struct {
	UserID uuid.UUID
}
