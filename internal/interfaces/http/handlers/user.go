// internal/interfaces/http/handlers/user.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/list_users"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	authmiddleware "github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type UserHandler struct {
	getUser    *get_user.Handler
	listUsers  *list_users.Handler
	deactivate *deactivate_user.Handler
	activate   *activate_user.Handler
	templates  *Templates
	logger     ports.Logger
}

func NewUserHandler(
	getUser *get_user.Handler,
	listUsers *list_users.Handler,
	deactivate *deactivate_user.Handler,
	activate *activate_user.Handler,
	templates *Templates,
	logger ports.Logger,
) *UserHandler {
	return &UserHandler{
		getUser:    getUser,
		listUsers:  listUsers,
		deactivate: deactivate,
		activate:   activate,
		templates:  templates,
		logger:     logger,
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	view, err := h.getUser.Handle(r.Context(), get_user.Query{UserID: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.writeJSON(w, http.StatusOK, view)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	view, err := h.listUsers.Handle(r.Context(), list_users.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, http.StatusOK, view)
}

// ListUsersPage — публичный HTML-маршрут для просмотра списка пользователей
// userPageData — набор полей, который ожидает шаблон user.html.
type userPageData struct {
	IsAdmin         bool
	UserID          uuid.UUID
	StatusLabel     string
	LastName        string
	FirstName       string
	MiddleName      string
	BirthDate       string
	HiredAt         string
	District        string
	Department      string
	WorkdayDuration string
	Email           string
	IsInvalid       bool
	Experience      string
	Initials        string
	PositionTitle   string

	YearStats []yearStatView
	Current   currentYearView
	Warnings  []string
}

type yearStatView struct {
	From      string
	To        string
	Base      int
	Seniority int
	Irregular int
	Shifts    []shiftView
}

type shiftView struct {
	From   string
	To     string
	Reason string
}

type currentYearView struct {
	Range     string
	Total     int
	Earned    float64
	Used      int
	Available int
	Remaining float64
	Carryover int
	Today     string
}

// assembleUserPageData собирает flat-слайс данных из domain-моделей для шаблонa.
func assembleUserPageData(result *get_user.Result) *userPageData {
	u := result.User

	data := &userPageData{
		UserID:          u.ID().UUID(),
		StatusLabel:     u.Status().Label(),
		LastName:        u.LastName(),
		FirstName:       u.FirstName(),
		MiddleName:      u.MiddleName(),
		BirthDate:       u.BirthDate().String(),
		HiredAt:         u.HiredAt().String(),
		District:        u.District().Title(),
		Department:      u.Department().String(),
		WorkdayDuration: fmt.Sprintf("%d ч.", u.WorkdayDuration()),
		Email:           u.Email().Value(),
		IsInvalid:       u.IsInvalid(),
		Experience:      u.ExperienceLabel(),
		Initials:        u.Initials(),
		Warnings:        result.Warnings,
	}

	if pos := u.Position(); pos != "" {
		data.PositionTitle = pos.Title()
	}

	// Year stats with shifts
	for _, ys := range result.YearStats {
		sv := yearStatView{
			From:      ys.From.Format("02.01.2006"),
			To:        ys.To.AddDate(0, 0, -1).Format("02.01.2006"), // To exclusive in model
			Base:      ys.Base,
			Seniority: ys.Seniority,
			Irregular: ys.Irregular,
			Shifts:    make([]shiftView, 0),
		}
		for _, s := range ys.Shifts {
			name, _ := vacation.ShiftKindName(s.Kind)
			sv.Shifts = append(sv.Shifts, shiftView{
				From:   s.From.Format("02.01.2006"),
				To:     s.To.Format("02.01.2006"),
				Reason: name,
			})
		}
		data.YearStats = append(data.YearStats, sv)
	}

	// Current year info
	if c := result.Current; !isZeroCurrent(c) {
		todayInYear := c.Year.From
		if todayInYear.After(c.Today) {
			todayInYear = todayInYear.Truncate(24 * time.Hour)
		}
		totalEntitlement := c.Year.Base + c.Year.Seniority + c.Year.Irregular
		data.Current = currentYearView{
			Range:     fmt.Sprintf("%s – %s", c.Year.From.Format("02.01.2006"), c.Year.To.AddDate(0, 0, -1).Format("02.01.2006")),
			Total:     totalEntitlement,
			Earned:    c.Earned,
			Used:      c.Used,
			Available: totalEntitlement - c.Used,
			Remaining: c.Remaining,
			Carryover: result.Carryover,
			Today:     c.Today.Format("02.01.2006"),
		}
	}

	return data
}

// isZeroCurrent проверяет, пустая ли CurrentYearInfo.
func isZeroCurrent(c vacation.CurrentYearInfo) bool {
	return c.Year.From.IsZero()
}

func (h *UserHandler) ListUsersPage(w http.ResponseWriter, r *http.Request) {
	result, err := h.listUsers.Handle(r.Context(), list_users.Query{})
	if err != nil {
		h.logger.Error(r.Context(), "list users failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Users": result.Users,
		"Total": result.Total,
		"Page":  result.Page,
		"Size":  result.Size,
		"Filters": struct {
			LastName, FirstName, MiddleName, Email, Status, BirthDate, HiredAt string
		}{},
	}

	if err := h.templates.Render(w, "users.html", data); err != nil {
		h.logger.Error(r.Context(), "render users page failed", "error", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// GetUserPage — публичный HTML-маршрут для просмотра карточки пользователя
func (h *UserHandler) GetUserPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.getUser.Handle(r.Context(), get_user.Query{UserID: id})
	if err != nil {
		h.logger.Error(r.Context(), "get user failed", "user_id", id, "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := assembleUserPageData(result)
	currentUser, _ := authmiddleware.UserFromContext(r.Context())
	data.IsAdmin = currentUser.IsAdmin()

	if err := h.templates.Render(w, "user.html", data); err != nil {
		h.logger.Error(r.Context(), "render user page failed", "error", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.deactivate.Handle(r.Context(), deactivate_user.Command{UserID: id}); err != nil {
		h.logger.Error(context.Background(), "deactivate user failed", "user_id", id.String(), "error", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.activate.Handle(r.Context(), activate_user.Command{UserID: id}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error(context.Background(), "writeJSON: encode failed", "error", err)
	}
}

func (h *UserHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}
