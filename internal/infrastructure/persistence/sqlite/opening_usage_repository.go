package sqlite

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/jmoiron/sqlx"
)

type openingUsageRepository struct {
	db *sqlx.DB
}

var _ ports.OpeningUsageRepository = (*openingUsageRepository)(nil)

func NewOpeningUsageRepository(db *sqlx.DB) *openingUsageRepository {
	return &openingUsageRepository{db: db}
}

func (r *openingUsageRepository) ListByUser(ctx context.Context, userID user.UserID) ([]vacation.OpeningUsage, error) {
	var rows []struct {
		WorkYearStart string `db:"work_year_start"`
		Base          int    `db:"base"`
		Irregular     int    `db:"irregular"`
		Seniority     int    `db:"seniority"`
	}
	const q = `SELECT work_year_start, base, irregular, seniority
		FROM vacation_opening_usage WHERE user_id = ? ORDER BY work_year_start`
	if err := r.db.SelectContext(ctx, &rows, q, userID.String()); err != nil {
		return nil, fmt.Errorf("list opening usage: %w", err)
	}

	out := make([]vacation.OpeningUsage, 0, len(rows))
	for _, row := range rows {
		start, err := date.ParseDate(row.WorkYearStart)
		if err != nil {
			return nil, fmt.Errorf("opening usage date %q: %w", row.WorkYearStart, err)
		}
		out = append(out, vacation.OpeningUsage{
			WorkYearStart: start,
			Days:          vacation.OpeningDays{Base: row.Base, Irregular: row.Irregular, Seniority: row.Seniority},
		})
	}
	return out, nil
}

// Replace заменяет все записи сотрудника одной транзакцией.
func (r *openingUsageRepository) Replace(ctx context.Context, userID user.UserID, usages []vacation.OpeningUsage) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM vacation_opening_usage WHERE user_id = ?`, userID.String()); err != nil {
		return fmt.Errorf("clear opening usage: %w", err)
	}
	const q = `INSERT INTO vacation_opening_usage (user_id, work_year_start, base, irregular, seniority) VALUES (?, ?, ?, ?, ?)`
	for _, u := range usages {
		if _, err := tx.ExecContext(ctx, q, userID.String(), u.WorkYearStart.ISO(), u.Days.Base, u.Days.Irregular, u.Days.Seniority); err != nil {
			return fmt.Errorf("save opening usage: %w", err)
		}
	}
	return tx.Commit()
}
