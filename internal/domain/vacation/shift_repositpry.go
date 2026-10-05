package vacation

import (
	"context"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type ShiftRepository interface {
	ListByUser(ctx context.Context, userID user.UserID) ([]Shift, error)
	ListByUserInRange(ctx context.Context, userID user.UserID, from, to time.Time) ([]Shift, error)
	Save(ctx context.Context, userID user.UserID, shift Shift) error
}