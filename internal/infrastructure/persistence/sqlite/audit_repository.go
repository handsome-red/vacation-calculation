package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/audit"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/jmoiron/sqlx"
)

type auditRepository struct {
	db *sqlx.DB
}

var _ ports.AuditRepository = (*auditRepository)(nil)

func NewAuditRepository(db *sqlx.DB) *auditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Record(ctx context.Context, e audit.Entry) error {
	const q = `
		INSERT INTO audit_log (at, actor_id, actor_name, target_id, target_name, district_code, action, details)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, e.At.UTC().Format(time.RFC3339), e.ActorID, e.ActorName,
		e.TargetID, e.TargetName, e.DistrictCode, e.Action, e.Details)
	if err != nil {
		return fmt.Errorf("record audit entry: %w", err)
	}
	return nil
}

func (r *auditRepository) List(ctx context.Context, f ports.AuditFilter) ([]audit.Entry, error) {
	var (
		conditions []string
		args       []any
	)
	if f.TargetID != "" {
		conditions = append(conditions, "target_id = ?")
		args = append(args, f.TargetID)
	}
	if f.DistrictCodes != nil {
		if len(f.DistrictCodes) == 0 {
			return nil, nil
		}
		marks := strings.TrimSuffix(strings.Repeat("?,", len(f.DistrictCodes)), ",")
		conditions = append(conditions, "district_code IN ("+marks+")")
		for _, d := range f.DistrictCodes {
			args = append(args, d)
		}
	}
	q := `SELECT at, actor_id, actor_name, target_id, target_name, district_code, action, details FROM audit_log`
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " ORDER BY id DESC LIMIT ?"
	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}
	args = append(args, limit)

	var rows []struct {
		At           string `db:"at"`
		ActorID      string `db:"actor_id"`
		ActorName    string `db:"actor_name"`
		TargetID     string `db:"target_id"`
		TargetName   string `db:"target_name"`
		DistrictCode string `db:"district_code"`
		Action       string `db:"action"`
		Details      string `db:"details"`
	}
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("list audit entries: %w", err)
	}
	out := make([]audit.Entry, 0, len(rows))
	for _, row := range rows {
		at, err := time.Parse(time.RFC3339, row.At)
		if err != nil {
			return nil, fmt.Errorf("audit entry time %q: %w", row.At, err)
		}
		out = append(out, audit.Entry{
			At: at, ActorID: row.ActorID, ActorName: row.ActorName,
			TargetID: row.TargetID, TargetName: row.TargetName, DistrictCode: row.DistrictCode,
			Action: row.Action, Details: row.Details,
		})
	}
	return out, nil
}
