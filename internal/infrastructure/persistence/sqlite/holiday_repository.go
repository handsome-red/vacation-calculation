package sqlite

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/jmoiron/sqlx"
)

type holidayRepository struct {
	db *sqlx.DB
}

var _ ports.HolidayRepository = (*holidayRepository)(nil)

func NewHolidayRepository(db *sqlx.DB) *holidayRepository {
	return &holidayRepository{
		db: db,
	}
}

func (r *holidayRepository) Save(ctx context.Context, h vacation.Holiday) error {
	if h.Date().IsZero() {
		return vacation.ErrHolidayDateEmpty
	}

	const q = `
		INSERT INTO holidays (date, name)
		VALUES (?, ?)
		ON CONFLICT (date) DO UPDATE SET name = EXCLUDED.name
	`

	_, err := r.db.ExecContext(ctx, q, h.Date().ISO(), h.Name())
	if err != nil {
		return fmt.Errorf("save holiday %q: %w", h.Date().ISO(), err)
	}

	return nil
}

func (r *holidayRepository) ListInRange(ctx context.Context, from, to vacation.Date) ([]vacation.Holiday, error) {
	const q = `
		SELECT date, name
		FROM holidays
		WHERE date BETWEEN ? AND ?
		ORDER BY date
	`

	var holidaysRows []holidayRow
	err := r.db.SelectContext(ctx, &holidaysRows, q, from, to)
	if err != nil {
		return nil, fmt.Errorf("list holidays: %w", err)
	}

	return rowsToHoliday(holidaysRows)
}

func (r *holidayRepository) ListByYear(ctx context.Context, year int) ([]vacation.Holiday, error) {
	const q = `
		SELECT name, date
		FROM holidays
		WHERE date >= ? AND date <= ?
		ORDER BY date
	`

	from := fmt.Sprintf("%04d-01-01", year)
	to := fmt.Sprintf("%04d-12-31", year)

	var holidaysRows []holidayRow
	if err := r.db.SelectContext(ctx, &holidaysRows, q, from, to); err != nil {
		return nil, fmt.Errorf("list holidays by year: %w", err)
	}

	return rowsToHoliday(holidaysRows)
}

func rowsToHoliday(holidaysRows []holidayRow) ([]vacation.Holiday, error) {
	result := make([]vacation.Holiday, 0, len(holidaysRows))
	for _, row := range holidaysRows {
		h, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("holiday row %q: %w", row.Date, err)
		}
		result = append(result, h)
	}

	return result, nil
}
