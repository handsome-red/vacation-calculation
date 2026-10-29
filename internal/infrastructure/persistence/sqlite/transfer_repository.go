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

type transferRepository struct {
	db *sqlx.DB
}

var _ ports.TransferRepository = (*transferRepository)(nil)

func NewTransferRepository(db *sqlx.DB) *transferRepository {
	return &transferRepository{db: db}
}

type transferRow struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	VacationID string    `db:"vacation_id"`
	OldStart   date.Date `db:"old_start"`
	OldEnd     date.Date `db:"old_end"`
	NewStart   date.Date `db:"new_start"`
	NewEnd     date.Date `db:"new_end"`
	Reason     string    `db:"reason"`
	CreatedBy  string    `db:"created_by"`
	CreatedAt  string    `db:"created_at"`
}

func (r *transferRepository) Save(ctx context.Context, t vacation.Transfer) error {
	const q = `
		INSERT INTO vacation_transfers (
			id, user_id, vacation_id, old_start, old_end, new_start, new_end,
			reason, created_by, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q,
		t.ID(), t.UserID().String(), t.VacationID().String(),
		t.OldStart().ISO(), t.OldEnd().ISO(), t.NewStart().ISO(), t.NewEnd().ISO(),
		t.Reason(), t.CreatedBy().String(), t.CreatedAt().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return fmt.Errorf("save vacation transfer: %w", err)
	}
	return nil
}

func (r *transferRepository) ListByUser(ctx context.Context, userID user.UserID) ([]vacation.Transfer, error) {
	const q = `
		SELECT id, user_id, vacation_id, old_start, old_end, new_start, new_end,
		       reason, created_by, created_at
		FROM vacation_transfers
		WHERE user_id = ?
		ORDER BY created_at
	`
	var rows []transferRow
	if err := r.db.SelectContext(ctx, &rows, q, userID.String()); err != nil {
		return nil, fmt.Errorf("list vacation transfers: %w", err)
	}

	out := make([]vacation.Transfer, 0, len(rows))
	for _, row := range rows {
		uid, err := user.ParseUserID(row.UserID)
		if err != nil {
			return nil, fmt.Errorf("transfer %q: user id: %w", row.ID, err)
		}
		vid, err := vacation.ParseVacationID(row.VacationID)
		if err != nil {
			return nil, fmt.Errorf("transfer %q: vacation id: %w", row.ID, err)
		}
		by, err := user.ParseUserID(row.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("transfer %q: created_by: %w", row.ID, err)
		}
		createdAt, err := parseSQLiteTime(row.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("transfer %q: created_at: %w", row.ID, err)
		}
		out = append(out, vacation.ReconstructTransfer(vacation.TransferParams{
			ID: row.ID, UserID: uid, VacationID: vid,
			OldStart: row.OldStart, OldEnd: row.OldEnd, NewStart: row.NewStart, NewEnd: row.NewEnd,
			Reason: row.Reason, CreatedBy: by, CreatedAt: createdAt,
		}))
	}
	return out, nil
}
