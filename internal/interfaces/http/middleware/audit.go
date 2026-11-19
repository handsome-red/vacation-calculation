package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/handsome-red/vacation-calculation/internal/domain/audit"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

// AuditAction — изменяющий маршрут, который попадает в журнал.
type AuditAction struct {
	Method  string
	Pattern string // полный шаблон маршрута chi, например /users/{id}/edit
	Action  string
	// TargetParam — параметр маршрута с ID сотрудника; пусто — сотрудника нет или его
	// указывает обработчик через AuditTarget.
	TargetParam string
	// Describe — подробности после успешной обработки; nil — без подробностей.
	Describe func(r *http.Request) string
}

// Auditor записывает в журнал успешные изменяющие запросы из списка действий.
// Применяется после Require: действующий пользователь берётся из контекста.
type Auditor struct {
	users   ports.UserRepository
	repo    ports.AuditRepository
	logger  ports.Logger
	actions map[string]AuditAction
}

func NewAuditor(users ports.UserRepository, repo ports.AuditRepository, logger ports.Logger, actions []AuditAction) *Auditor {
	m := make(map[string]AuditAction, len(actions))
	for _, a := range actions {
		m[a.Method+" "+a.Pattern] = a
	}
	return &Auditor{users: users, repo: repo, logger: logger, actions: m}
}

type auditNoteKey struct{}

// auditNote — то, что обработчик сообщает журналу сам: подробности, которые не
// восстановить после обработки (тело JSON), и созданного сотрудника.
type auditNote struct {
	details  string
	targetID string
}

// AuditDetail задаёт подробности записи журнала для текущего запроса.
func AuditDetail(ctx context.Context, details string) {
	if n, ok := ctx.Value(auditNoteKey{}).(*auditNote); ok {
		n.details = details
	}
}

// AuditTarget задаёт сотрудника записи журнала, если его нет в маршруте (новый сотрудник).
func AuditTarget(ctx context.Context, userID string) {
	if n, ok := ctx.Value(auditNoteKey{}).(*auditNote); ok {
		n.targetID = userID
	}
}

func (a *Auditor) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}
		note := &auditNote{}
		r = r.WithContext(context.WithValue(r.Context(), auditNoteKey{}, note))
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		if status := ww.Status(); status >= 400 {
			return
		}
		rctx := chi.RouteContext(r.Context())
		if rctx == nil {
			return
		}
		action, ok := a.actions[r.Method+" "+rctx.RoutePattern()]
		if !ok {
			return
		}
		a.record(r, action, note)
	})
}

func (a *Auditor) record(r *http.Request, action AuditAction, note *auditNote) {
	ctx := r.Context()
	actor, ok := UserFromContext(ctx)
	if !ok || actor == nil {
		return
	}
	e := audit.Entry{
		At:        time.Now(),
		ActorID:   actor.ID().String(),
		ActorName: actor.Initials(),
		Action:    action.Action,
		Details:   note.details,
	}
	if action.Describe != nil {
		if d := action.Describe(r); d != "" {
			e.Details = strings.TrimSpace(strings.Join([]string{d, e.Details}, " "))
		}
	}

	targetID := note.targetID
	if action.TargetParam != "" {
		targetID = chi.URLParam(r, action.TargetParam)
	}
	if targetID != "" {
		if id, err := user.ParseUserID(targetID); err == nil {
			if target, err := a.users.FindByID(ctx, id); err == nil {
				e.TargetID = target.ID().String()
				e.TargetName = target.Initials()
				e.DistrictCode = target.District().Code()
			}
		}
	}
	if e.TargetID == "" {
		// Действие без сотрудника (график теротдела) относится к выбранному теротделу.
		e.DistrictCode = r.FormValue("district")
	}

	// Запрос уже выполнен: ошибка журнала не отменяет действие, только пишется в лог.
	if err := a.repo.Record(context.WithoutCancel(ctx), e); err != nil {
		a.logger.Error(ctx, "audit record failed", "error", err, "action", action.Action)
	}
}
