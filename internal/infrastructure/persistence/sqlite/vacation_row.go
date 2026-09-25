// infrastructure/persistence/sqlite/vacation_row.go
package sqlite

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type vacationRow struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	StartDate string `db:"start_date"`
	EndDate   string `db:"end_date"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (r vacationRow) toDomain() (*vacation.Vacation, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid vacation id: %w", err)
	}

	vacationID, err := vacation.VacationIDFromUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid vacation id: %w", err)
	}

	userID, err := user.ParseUserID(r.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	startDate, err := vacation.ParseDate(r.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	endDate, err := vacation.ParseDate(r.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at: %w", err)
	}

	return vacation.ReconstructVacation(vacation.VacationParams{
		ID:        vacationID,
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		CreatedAt: createdAt.UTC(),
		UpdatedAt: updatedAt.UTC(),
	}), nil
}