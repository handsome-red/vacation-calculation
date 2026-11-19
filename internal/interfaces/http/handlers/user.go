// internal/interfaces/http/handlers/user.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/update_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/edit_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/list_users"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	authmiddleware "github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type UserHandler struct {
	getUser    *get_user.Handler
	listUsers  *list_users.Handler
	deactivate *deactivate_user.Handler
	activate   *activate_user.Handler
	update     *update_user.Handler
	editForm   *edit_form.Handler
	templates  *Templates
	logger     ports.Logger
}

func NewUserHandler(
	getUser *get_user.Handler,
	listUsers *list_users.Handler,
	deactivate *deactivate_user.Handler,
	activate *activate_user.Handler,
	update *update_user.Handler,
	editForm *edit_form.Handler,
	templates *Templates,
	logger ports.Logger,
) *UserHandler {
	return &UserHandler{
		getUser:    getUser,
		listUsers:  listUsers,
		deactivate: deactivate,
		activate:   activate,
		update:     update,
		editForm:   editForm,
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
	caller, _ := authmiddleware.CallerFromContext(r.Context())
	view, err := h.listUsers.Handle(r.Context(), list_users.Query{Caller: caller})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, http.StatusOK, view)
}

// ListUsersPage — публичный HTML-маршрут для просмотра списка пользователей
// userPageData — набор полей, который ожидает шаблон user.html.
type userPageData struct {
	// CanManage — смотрящий управляет сотрудником: администратор или руководитель его теротдела.
	CanManage       bool
	Role            string
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
		Role:            u.RoleLabel(),
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
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))
	filters := list_users.Filters{
		LastName:       strings.TrimSpace(q.Get("last_name")),
		FirstName:      strings.TrimSpace(q.Get("first_name")),
		MiddleName:     strings.TrimSpace(q.Get("middle_name")),
		Email:          strings.TrimSpace(q.Get("email")),
		Status:         q.Get("status"),
		Category:       q.Get("category"),
		Position:       q.Get("position"),
		DistrictCode:   q.Get("district"),
		DepartmentCode: q.Get("department"),
	}

	caller, _ := authmiddleware.CallerFromContext(r.Context())
	result, err := h.listUsers.Handle(r.Context(), list_users.Query{Page: page, Size: size, Filters: filters, Caller: caller})
	if err != nil {
		h.logger.Error(r.Context(), "list users failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ссылки на страницы сохраняют фильтры.
	pageURL := func(p int) string {
		v := r.URL.Query()
		v.Set("page", strconv.Itoa(p))
		return "/users?" + v.Encode()
	}
	prevURL, nextURL := "", ""
	if result.Page > 1 {
		prevURL = pageURL(result.Page - 1)
	}
	if result.Page < result.Pages {
		nextURL = pageURL(result.Page + 1)
	}

	data := map[string]any{
		"Users":       result.Users,
		"Total":       result.Total,
		"Page":        result.Page,
		"Pages":       result.Pages,
		"Size":        result.Size,
		"PrevURL":     prevURL,
		"NextURL":     nextURL,
		"Filters":     filters,
		"Districts":   result.Districts,
		"Departments": result.Departments,
		"Positions":   result.Positions,
		"Categories":  result.Categories,
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
	data.CanManage = currentUser.Manages(result.User)

	if err := h.templates.Render(w, "user.html", data); err != nil {
		h.logger.Error(r.Context(), "render user page failed", "error", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// EditForm — форма изменения данных сотрудника (администратор или руководитель его теротдела).
func (h *UserHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	caller, _ := authmiddleware.CallerFromContext(r.Context())
	view, err := h.editForm.Handle(r.Context(), edit_form.Query{UserID: chi.URLParam(r, "id"), Caller: caller})
	if err != nil {
		h.renderEditError(w, r, err)
		return
	}
	if err := h.templates.Render(w, "user_edit.html", view); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// UpdateUser сохраняет данные сотрудника; при ошибке ввода показывает форму с введёнными значениями.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	// Нечисловой стаж и рабочий день отклонит use case.
	experience, err := strconv.Atoi(r.FormValue("experience"))
	if err != nil {
		experience = -1
	}
	workday, err := strconv.Atoi(r.FormValue("workday_duration"))
	if err != nil {
		workday = 0
	}
	cmd := update_user.Command{
		UserID:          id,
		LastName:        strings.TrimSpace(r.FormValue("last_name")),
		FirstName:       strings.TrimSpace(r.FormValue("first_name")),
		MiddleName:      strings.TrimSpace(r.FormValue("middle_name")),
		BirthDate:       r.FormValue("birth_date"),
		Position:        r.FormValue("position"),
		HiredAt:         r.FormValue("hired_at"),
		Experience:      experience,
		DepartmentCode:  r.FormValue("department"),
		DistrictCode:    r.FormValue("district"),
		WorkdayDuration: workday,
		Email:           r.FormValue("email"),
		IsInvalid:       r.FormValue("is_invalid") == "true",
		Password:        r.FormValue("password"),
	}
	cmd.Caller, _ = authmiddleware.CallerFromContext(r.Context())

	err = h.update.Handle(r.Context(), cmd)
	if err == nil {
		admin, _ := authmiddleware.UserFromContext(r.Context())
		h.logger.Info(r.Context(), "user profile updated", "user_id", id, "admin_id", admin.ID().String())
		http.Redirect(w, r, "/users/"+id, http.StatusSeeOther)
		return
	}
	if !errors.Is(err, update_user.ErrInvalidInput) {
		h.renderEditError(w, r, err)
		return
	}

	view, ferr := h.editForm.Handle(r.Context(), edit_form.Query{UserID: id, Caller: cmd.Caller})
	if ferr != nil {
		h.renderEditError(w, r, ferr)
		return
	}
	view.Form = edit_form.Form{
		LastName: cmd.LastName, FirstName: cmd.FirstName, MiddleName: cmd.MiddleName,
		Email: cmd.Email, BirthDate: cmd.BirthDate, Position: cmd.Position, HiredAt: cmd.HiredAt,
		Experience: r.FormValue("experience"), District: cmd.DistrictCode, Department: cmd.DepartmentCode,
		WorkdayDuration: r.FormValue("workday_duration"), IsInvalid: cmd.IsInvalid,
	}
	view.Error = err.Error()
	w.WriteHeader(http.StatusBadRequest)
	if err := h.templates.Render(w, "user_edit.html", view); err != nil {
		h.logger.Error(r.Context(), "render user edit page failed", "error", err)
	}
}

func (h *UserHandler) renderEditError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, user.ErrUserNotFound) {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}
	if errors.Is(err, update_user.ErrForbidden) {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}
	h.logger.Error(r.Context(), "edit user failed", "error", err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
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
