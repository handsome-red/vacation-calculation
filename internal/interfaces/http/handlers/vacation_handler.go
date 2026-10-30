package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/approve_vacation"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/change_status"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/create_vacations"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/export_report"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_vacations"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_vacation_form"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
	authmiddleware "github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type quotaDTO struct {
	Total   int `json:"total"`
	Blocked int `json:"blocked"`
}
type yearStatDTO struct {
	Year      string   `json:"year"`
	Start     string   `json:"start"`
	Base      quotaDTO `json:"base"`
	Seniority quotaDTO `json:"seniority"`
	Irregular quotaDTO `json:"irregular"`
}
type appData struct {
	UserID        string        `json:"userId"`
	HasActiveYear bool          `json:"hasActiveYear"`
	Available     int           `json:"available"`
	YearStats     []yearStatDTO `json:"yearStats"`
}

type vacationStatusOption struct {
	Code  string
	Label string
}

type VacationHandler struct {
	createVacationUseCase   *create_vacations.Handler
	approveVacationUseCase  *approve_vacation.Handler
	changeStatusUseCase     *change_status.Handler
	getUserVacationsUseCase *get_user_vacations.Handler
	getVacationFormUseCase  *get_vacation_form.Handler
	exportReportUseCase     *export_report.Handler
	templates               *Templates
	logger                  ports.Logger
}

func NewVacationHandler(
	createVacationUseCase *create_vacations.Handler,
	approveVacationUseCase *approve_vacation.Handler,
	changeStatusUseCase *change_status.Handler,
	getUserVacationsUseCase *get_user_vacations.Handler,
	getVacationFormUseCase *get_vacation_form.Handler,
	exportReportUseCase *export_report.Handler,
	templates *Templates,
	logger ports.Logger,
) *VacationHandler {
	return &VacationHandler{
		createVacationUseCase:   createVacationUseCase,
		approveVacationUseCase:  approveVacationUseCase,
		changeStatusUseCase:     changeStatusUseCase,
		getUserVacationsUseCase: getUserVacationsUseCase,
		getVacationFormUseCase:  getVacationFormUseCase,
		exportReportUseCase:     exportReportUseCase,
		templates:               templates,
		logger:                  logger,
	}
}

func (h *VacationHandler) GetVacations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("userId")
	userID, err := user.ParseUserID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	showAll := r.URL.Query().Get("all") == "1"
	result, err := h.getUserVacationsUseCase.Handle(r.Context(), get_user_vacations.Query{
		UserID: userID.String(),
		All:    showAll,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	currentUser, _ := authmiddleware.UserFromContext(r.Context())
	data := map[string]any{
		"IsAdmin":   currentUser.IsAdmin(),
		"UserID":    userID.String(),
		"Vacations": result.Vacations,
		"Hidden":    result.Hidden,
		"ShowAll":   showAll,
		"Statuses": []vacationStatusOption{
			{vacation.StatusDraft.ISO(), vacation.StatusDraft.String()},
			{vacation.StatusApproved.ISO(), vacation.StatusApproved.String()},
			{vacation.StatusCancelled.ISO(), vacation.StatusCancelled.String()},
		},
	}

	if err := h.templates.Render(w, "vacations.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *VacationHandler) CreateVacation(w http.ResponseWriter, r *http.Request) {
	caller, ok := middleware.CallerFromContext(r.Context())
	if !ok {
		http.Error(w, "Необходим вход", http.StatusUnauthorized)
		return
	}

	var cmd create_vacations.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.createVacationUseCase.Handle(r.Context(), cmd, caller)
	if err != nil {
		http.Error(w, err.Error(), vacationErrorStatus(err))
		return
	}

	if len(result.Warnings) > 0 {
		h.logger.Warn(r.Context(), "vacation created by admin bypassing rules",
			"user_id", cmd.UserID, "admin_id", caller.UserID.String(), "warnings", result.Warnings)
	}
	h.logger.Info(r.Context(), "vacation created successfully")
	http.Redirect(w, r, "/users/"+cmd.UserID, http.StatusSeeOther)
}

type allocationDTO struct {
	Start         string `json:"start"`
	End           string `json:"end"`
	Kind          string `json:"kind"`
	KindLabel     string `json:"kindLabel"`
	WorkYearStart string `json:"workYearStart"`
	Days          int    `json:"days"`
}

// PreviewVacation возвращает распределение выбранных периодов по квотам рабочих лет,
// ничего не сохраняя. Ошибки правил отдаются как {"error": "..."} с кодом 400.
func (h *VacationHandler) PreviewVacation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	caller, ok := middleware.CallerFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Необходим вход"})
		return
	}

	var cmd create_vacations.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	result, err := h.createVacationUseCase.Preview(r.Context(), cmd, caller)
	if err != nil {
		w.WriteHeader(vacationErrorStatus(err))
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	out := make([]allocationDTO, 0, len(result.Allocations))
	for _, a := range result.Allocations {
		out = append(out, allocationDTO{
			Start:         a.Start.ISO(),
			End:           a.End.ISO(),
			Kind:          string(a.Kind),
			KindLabel:     a.Kind.Label(),
			WorkYearStart: a.WorkYearStart.ISO(),
			Days:          a.Days,
		})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"allocations": out, "warnings": result.Warnings, "notices": result.Notices})
}

func vacationErrorStatus(err error) int {
	switch {
	case errors.Is(err, vacation.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, vacation.ErrInvalidInput), errors.Is(err, vacation.ErrBasicVacationRules):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *VacationHandler) GetVacationForm(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	result, err := h.getVacationFormUseCase.Handle(r.Context(), get_vacation_form.Query{
		UserID: userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	yearStats := make([]yearStatDTO, len(result.YearStats))
	for i, y := range result.YearStats {
		yearStats[i] = yearStatDTO{
			Year:      y.Year,
			Start:     y.Start,
			Base:      quotaDTO{Total: y.Base.Total, Blocked: y.Base.Blocked},
			Seniority: quotaDTO{Total: y.Seniority.Total, Blocked: y.Seniority.Blocked},
			Irregular: quotaDTO{Total: y.Irregular.Total, Blocked: y.Irregular.Blocked},
		}
	}

	raw, err := json.Marshal(appData{
		UserID:        result.UserID,
		HasActiveYear: result.HasActiveYear,
		Available:     result.Available,
		YearStats:     yearStats,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	year := now.Year()
	if y, err := strconv.Atoi(r.URL.Query().Get("year")); err == nil && y >= 2000 && y <= 2100 {
		year = y
	}
	currentUser, _ := authmiddleware.UserFromContext(r.Context())
	data := map[string]any{
		"Year":    year,
		"IsAdmin": currentUser != nil && currentUser.IsAdmin(),
		"Cmd":     create_vacations.Command{UserID: result.UserID},
		"Error":   "",
		"Calendar": CalendarProps{
			UserID:    result.UserID,
			Year:      now.Year(),
			Month:     int(now.Month()),
			FromField: "start_date",
			ToField:   "end_date",
		},
		"Balance": struct {
			HasActiveYear bool
			Available     int
		}{result.HasActiveYear, result.Available},
		"AppDataJSON": template.JS(raw),
		"Warnings":    result.Warnings,
	}

	if err := h.templates.Render(w, "vacation_form.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *VacationHandler) ApproveVacation(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	vacationID := r.PathValue("id")

	cmd := approve_vacation.Command{
		UserID:     userID,
		VacationID: vacationID,
	}

	_, err := h.approveVacationUseCase.Handle(r.Context(), cmd)
	if err != nil {
		h.logger.Error(r.Context(), "failed to approve vacation", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info(r.Context(), "vacation approved successfully")
	redirectURL := "/api/v1/users/" + userID + "/vacations"
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *VacationHandler) ChangeVacationStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	cmd := change_status.Command{
		UserID:     r.PathValue("userId"),
		VacationID: r.PathValue("id"),
		Status:     r.FormValue("status"),
	}
	if err := h.changeStatusUseCase.Handle(r.Context(), cmd); err != nil {
		h.logger.Error(r.Context(), "failed to change vacation status", "error", err)
		if errors.Is(err, vacation.ErrStatusInvalid) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/api/v1/users/"+cmd.UserID+"/vacations", http.StatusSeeOther)
}

func (h *VacationHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")

	if year == "" {
		http.Error(w, "не указан год", http.StatusBadRequest)
		return
	}

	data, err := h.exportReportUseCase.Handle(r.Context(), export_report.Query{
		Year:         0,
		DistrictCode: "",
	})
	if err != nil {
		h.logger.Info(r.Context(), "export vacation repor", "error", err)
		http.Error(w, "не удалось сформировать отчёт", http.StatusInternalServerError)
		return
	}

	// TODO
	_ = data
}
