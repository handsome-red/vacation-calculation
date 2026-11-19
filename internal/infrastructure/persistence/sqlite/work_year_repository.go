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

type workYearRepository struct {
	db *sqlx.DB
}

var _ ports.WorkYearRepository = (*workYearRepository)(nil)

func NewWorkYearRepository(db *sqlx.DB) *workYearRepository {
	return &workYearRepository{db: db}
}

func (r *workYearRepository) ListByUser(ctx context.Context, userID user.UserID) ([]vacation.WorkYear, error) {
	var rows []struct {
		Number    int    `db:"number"`
		StartDate string `db:"start_date"`
		EndDate   string `db:"end_date"`
		Manual    int    `db:"manual"`
	}
	const q = `SELECT number, start_date, end_date, manual FROM work_years WHERE user_id = ? ORDER BY number`
	if err := r.db.SelectContext(ctx, &rows, q, userID.String()); err != nil {
		return nil, fmt.Errorf("list work years: %w", err)
	}

	out := make([]vacation.WorkYear, 0, len(rows))
	for _, row := range rows {
		start, err := date.ParseDate(row.StartDate)
		if err != nil {
			return nil, fmt.Errorf("work year start %q: %w", row.StartDate, err)
		}
		end, err := date.ParseDate(row.EndDate)
		if err != nil {
			return nil, fmt.Errorf("work year end %q: %w", row.EndDate, err)
		}
		out = append(out, vacation.WorkYear{Number: row.Number, Start: start, End: end, Manual: row.Manual != 0})
	}
	return out, nil
}

func (r *workYearRepository) Add(ctx context.Context, userID user.UserID, years []vacation.WorkYear) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, y := range years {
		const q = `INSERT INTO work_years (user_id, number, start_date, end_date, manual) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (user_id, number) DO NOTHING`
		if _, err := tx.ExecContext(ctx, q, userID.String(), y.Number, y.Start.ISO(), y.End.ISO(), boolToInt(y.Manual)); err != nil {
			return fmt.Errorf("add work year: %w", err)
		}
	}
	return tx.Commit()
}

func (r *workYearRepository) Apply(
	ctx context.Context,
	userID user.UserID,
	years []vacation.WorkYear,
	vacations []vacation.VacationWorkYear,
	opening []vacation.OpeningUsage,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM work_years WHERE user_id = ?`, userID.String()); err != nil {
		return fmt.Errorf("clear work years: %w", err)
	}
	for _, y := range years {
		const q = `INSERT INTO work_years (user_id, number, start_date, end_date, manual) VALUES (?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(ctx, q, userID.String(), y.Number, y.Start.ISO(), y.End.ISO(), boolToInt(y.Manual)); err != nil {
			return fmt.Errorf("save work year: %w", err)
		}
	}

	for _, v := range vacations {
		const q = `UPDATE vacations SET work_year_start = ? WHERE id = ? AND user_id = ?`
		if _, err := tx.ExecContext(ctx, q, v.Start.ISO(), v.ID.String(), userID.String()); err != nil {
			return fmt.Errorf("move vacation: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM vacation_opening_usage WHERE user_id = ?`, userID.String()); err != nil {
		return fmt.Errorf("clear opening usage: %w", err)
	}
	for _, u := range opening {
		const q = `INSERT INTO vacation_opening_usage (user_id, work_year_start, base, irregular, seniority) VALUES (?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(ctx, q, userID.String(), u.WorkYearStart.ISO(), u.Days.Base, u.Days.Irregular, u.Days.Seniority); err != nil {
			return fmt.Errorf("save opening usage: %w", err)
		}
	}
	return tx.Commit()
}

func (r *workYearRepository) DeleteByUser(ctx context.Context, userID user.UserID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM work_years WHERE user_id = ?`, userID.String()); err != nil {
		return fmt.Errorf("delete work years: %w", err)
	}
	return nil
}
