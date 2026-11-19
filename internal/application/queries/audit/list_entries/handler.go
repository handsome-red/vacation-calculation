// Package list_entries — журнал изменений для администратора и руководителей.
package list_entries

import (
	"context"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

// limit — сколько последних записей показывается.
const limit = 300

// moscow — время в журнале показывается по Москве: сервер работает в UTC.
var moscow = time.FixedZone("МСК", 3*60*60)

type Query struct {
	Caller auth.Caller
	// TargetID — журнал одного сотрудника; пусто — общий журнал.
	TargetID string
}

type Row struct {
	At         string
	ActorID    string
	ActorName  string
	TargetID   string
	TargetName string
	Action     string
	Details    string
}

type Result struct {
	Rows  []Row
	Limit int
}

type Handler struct {
	repo ports.AuditRepository
}

func NewHandler(repo ports.AuditRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle возвращает последние записи. Руководитель видит только записи своих
// теротделов; общие действия (праздничные дни) — только администратор.
func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	entries, err := h.repo.List(ctx, ports.AuditFilter{
		TargetID:      q.TargetID,
		DistrictCodes: q.Caller.Scope(),
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}
	rows := make([]Row, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, Row{
			At:         e.At.In(moscow).Format("02.01.2006 15:04"),
			ActorID:    e.ActorID,
			ActorName:  e.ActorName,
			TargetID:   e.TargetID,
			TargetName: e.TargetName,
			Action:     e.Action,
			Details:    e.Details,
		})
	}
	return &Result{Rows: rows, Limit: limit}, nil
}
