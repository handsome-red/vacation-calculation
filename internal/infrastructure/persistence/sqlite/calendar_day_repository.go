package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/jmoiron/sqlx"
)

type calendarDayRepository struct {
	db *sqlx.DB
}

var _ ports.CalendarDayRepository = (*calendarDayRepository)(nil)

func NewCalendarDayRepository(db *sqlx.DB) *calendarDayRepository {
	return &calendarDayRepository{db: db}
}

func (r *calendarDayRepository) Save(ctx context.Context, d calendar.Exception) error {
	if d.Date().IsZero() {
		return vacation.ErrCalendarDayDateEmpty
	}

	const q = `
		INSERT INTO work_calendar_days (date, day_type, name)
		VALUES (?, ?, ?)
		ON CONFLICT (date) DO UPDATE SET
			day_type = EXCLUDED.day_type,
			name     = EXCLUDED.name
	`

	_, err := r.db.ExecContext(ctx, q, d.Date(), string(d.DayType()), d.Name())
	if err != nil {
		return fmt.Errorf("save calendar day %q: %w", d.Date().ISO(), err)
	}
	return nil
}

func (r *calendarDayRepository) ListInRange(ctx context.Context, from, to date.Date) ([]calendar.Exception, error) {
	const q = `
		SELECT date, day_type, name
		FROM work_calendar_days
		WHERE date BETWEEN ? AND ?
		ORDER BY date
	`

	var rows []calendarDayRow
	if err := r.db.SelectContext(ctx, &rows, q, from.ISO(), to.ISO()); err != nil {
		return nil, fmt.Errorf("list calendar days: %w", err)
	}
	return rowsToCalendarDays(rows)
}

func (r *calendarDayRepository) ListByYear(ctx context.Context, year int) ([]calendar.Exception, error) {
	from, err := date.NewDate(year, time.January, 1)
	if err != nil {
		return nil, fmt.Errorf("calendar year %d: %w", year, err)
	}
	to, err := date.NewDate(year, time.December, 31)
	if err != nil {
		return nil, fmt.Errorf("calendar year %d: %w", year, err)
	}
	return r.ListInRange(ctx, from, to)
}

func (r *calendarDayRepository) Delete(ctx context.Context, date date.Date) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM work_calendar_days WHERE date = ?`, date.ISO())
	if err != nil {
		return fmt.Errorf("delete calendar day %q: %w", date.ISO(), err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return vacation.ErrCalendarDayNotFound
	}
	return nil
}

func rowsToCalendarDays(rows []calendarDayRow) ([]calendar.Exception, error) {
	out := make([]calendar.Exception, 0, len(rows))
	for _, row := range rows {
		d, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("calendar day row %q: %w", row.Date, err)
		}
		out = append(out, d)
	}
	return out, nil
}
