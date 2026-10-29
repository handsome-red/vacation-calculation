package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type reminderRepository struct {
	db *sqlx.DB
}

var _ ports.ReminderRepository = (*reminderRepository)(nil)

func NewReminderRepository(db *sqlx.DB) *reminderRepository {
	return &reminderRepository{db: db}
}

func (r *reminderRepository) WasSent(ctx context.Context, userID user.UserID, start date.Date, recipient string) (bool, error) {
	var sent bool
	const q = `SELECT EXISTS (SELECT 1 FROM vacation_reminders WHERE user_id = ? AND start_date = ? AND recipient = ?)`
	if err := r.db.GetContext(ctx, &sent, q, userID.String(), start.ISO(), recipient); err != nil {
		return false, fmt.Errorf("check vacation reminder: %w", err)
	}
	return sent, nil
}

func (r *reminderRepository) MarkSent(ctx context.Context, userID user.UserID, start, end date.Date, recipient, email string, at time.Time) error {
	const q = `
		INSERT INTO vacation_reminders (user_id, start_date, end_date, recipient, email, sent_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id, start_date, recipient) DO NOTHING
	`
	if _, err := r.db.ExecContext(ctx, q, userID.String(), start.ISO(), end.ISO(), recipient, email, at.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("mark vacation reminder sent: %w", err)
	}
	return nil
}
