package vacation

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type VacationRepository interface {
	Save(ctx context.Context, vacation *Vacation) error
	// FindByUserID(ctx context.Context, userID UserID) ([]*Vacation, error)
	// FindByID(ctx context.Context, id VacationID) (*Vacation, error)
	FindByUserID(ctx context.Context, userID user.UserID) (*Vacation, error)
	// FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Vacation, error)
	// Delete(ctx context.Context, id VacationID) error
}
