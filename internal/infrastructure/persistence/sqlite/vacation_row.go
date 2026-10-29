// infrastructure/persistence/sqlite/vacation_row.go
package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type vacationRow struct {
	ID              string         `db:"id"`
	UserID          string         `db:"user_id"`
	StartDate       date.Date      `db:"start_date"`
	EndDate         date.Date      `db:"end_date"`
	Kind            string         `db:"kind"`
	WorkYearStart   date.Date      `db:"work_year_start"`
	TransferredFrom sql.NullString `db:"transferred_from"`
	Status          string         `db:"status"`
	CreatedAt       string         `db:"created_at"`
	UpdatedAt       string         `db:"updated_at"`
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

	startDate := r.StartDate

	endDate := r.EndDate

	kind, err := vacation.NewKind(r.Kind)
	if err != nil {
		return nil, fmt.Errorf("invalid kind: %w", err)
	}

	status, err := vacation.ParseStatus(r.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at: %w", err)
	}

	var transferredFrom vacation.VacationID
	if r.TransferredFrom.Valid {
		if transferredFrom, err = vacation.ParseVacationID(r.TransferredFrom.String); err != nil {
			return nil, fmt.Errorf("invalid transferred_from: %w", err)
		}
	}

	return vacation.ReconstructVacation(vacation.VacationParams{
		TransferredFrom: transferredFrom,
		ID:              vacationID,
		UserID:          userID,
		StartDate:       startDate,
		EndDate:         endDate,
		WorkYearStart:   r.WorkYearStart,
		Kind:            kind,
		VacationStatus:  status,
		CreatedAt:       createdAt.UTC(),
		UpdatedAt:       updatedAt.UTC(),
	}), nil
}
