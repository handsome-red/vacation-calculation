// infrastructure/persistence/sqlite/shift_repository.go
package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/jmoiron/sqlx"
)

type shiftRepository struct {
	db *sqlx.DB
}

func NewShiftRepository(db *sqlx.DB) vacation.ShiftRepository {
	return &shiftRepository{db: db}
}

var _ vacation.ShiftRepository = (*shiftRepository)(nil)

func (r *shiftRepository) ListByUser(ctx context.Context, userID user.UserID) ([]vacation.Shift, error) {
	const q = `
		SELECT kind, date_from, date_to
		FROM shifts
		WHERE user_id = ?
		ORDER BY date_from
	`

	var rows []struct {
		Kind     string `db:"kind"`
		DateFrom string `db:"date_from"`
		DateTo   string `db:"date_to"`
	}

	if err := r.db.SelectContext(ctx, &rows, q, userID.String()); err != nil {
		return nil, fmt.Errorf("list shifts: %w", err)
	}

	result := make([]vacation.Shift, 0, len(rows))
	for _, row := range rows {
		from, err := time.Parse("2006-01-02", row.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from %q: %w", row.DateFrom, err)
		}
		to, err := time.Parse("2006-01-02", row.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to %q: %w", row.DateTo, err)
		}

		result = append(result, vacation.Shift{
			Kind: vacation.ShiftKind(row.Kind),
			From: from,
			To:   to,
		})
	}
	return result, nil
}

func (r *shiftRepository)ListByUserInRange(ctx context.Context, userID user.UserID, from, to time.Time) ([]vacation.Shift, error) {
	return nil, nil
}
func (r *shiftRepository)    Save(ctx context.Context, userID user.UserID, shift vacation.Shift) error {
	return nil
}