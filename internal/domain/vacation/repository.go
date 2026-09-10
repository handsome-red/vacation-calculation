package vacation

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, vacation *Vacation) error
	// FindByUserID(ctx context.Context, userID UserID) ([]*Vacation, error)
	// FindByID(ctx context.Context, id VacationID) (*Vacation, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Vacation, error)
	// Delete(ctx context.Context, id VacationID) error
}
