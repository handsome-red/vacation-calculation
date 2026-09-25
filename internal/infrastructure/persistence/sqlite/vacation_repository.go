package sqlite

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/jmoiron/sqlx"
)

type vacationRepository struct {
	db *sqlx.DB
}

func NewVacationRepository(db *sqlx.DB) *vacationRepository{
	return &vacationRepository{
		db: db,
	}
}

func (r *vacationRepository) Save(ctx context.Context, v *vacation.Vacation) error {
	const q = `
		INSERT INTO vacations (
			id, user_id, start_date, end_date
		)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			user_id    = EXCLUDED.user_id,
			start_date = EXCLUDED.start_date,
			end_date   = EXCLUDED.end_date
	`

	_, err := r.db.ExecContext(
		ctx,
		q,
		v.ID().String(),
		v.UserID().String(),
		v.StartDate().Time().Format("2006-01-02"),
		v.EndDate().Time().Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("save vacation: %w", err)
	}

	return nil
}

func (r *vacationRepository) FindByUserID(ctx context.Context, userID user.UserID) ([]*vacation.Vacation, error) {
	const q = `
		SELECT id, user_id, start_date, end_date, created_at, updated_at
		FROM vacations
		WHERE user_id = ?
		ORDER BY start_date
	`

	var rows []vacationRow
	if err := r.db.SelectContext(ctx, &rows, q, userID.String()); err != nil {
		return nil, fmt.Errorf("find vacations by user: %w", err)
	}

	result := make([]*vacation.Vacation, 0, len(rows))
	for _, row := range rows {
		v, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("vacation row %q: %w", row.ID, err)
		}
		result = append(result, v)
	}

	return result, nil
}


// func(r *vacationRepository) NewVacationRepository()