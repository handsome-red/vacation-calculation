package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
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
	Base      quotaDTO `json:"base"`
	Seniority quotaDTO `json:"seniority"`
	Irregular quotaDTO `json:"irregular"`
}
type appData struct {
	UserID        string        `json:"userId"`
	HasActiveYear bool          `json:"hasActiveYear"`
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
	result, err := h.getUserVacationsUseCase.Handle(r.Context(), get_user_vacations.Query{
		UserID: userID.String(),
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

	_, err := h.createVacationUseCase.Handle(r.Context(), cmd, caller)
	if err != nil {
		if errors.Is(err, vacation.ErrInvalidInput) ||
			errors.Is(err, vacation.ErrBasicVacationRules) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info(r.Context(), "vacation created successfully")
	http.Redirect(w, r, "/users/"+cmd.UserID, http.StatusSeeOther)
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
			Base:      quotaDTO{Total: y.Base.Total, Blocked: y.Base.Blocked},
			Seniority: quotaDTO{Total: y.Seniority.Total, Blocked: y.Seniority.Blocked},
			Irregular: quotaDTO{Total: y.Irregular.Total, Blocked: y.Irregular.Blocked},
		}
	}

	raw, err := json.Marshal(appData{
		UserID:        result.UserID,
		HasActiveYear: result.HasActiveYear,
		YearStats:     yearStats,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	data := map[string]any{
		"Cmd":   create_vacations.Command{UserID: result.UserID},
		"Error": "",
		"Calendar": CalendarProps{
			UserID:    result.UserID,
			Year:      now.Year(),
			Month:     int(now.Month()),
			FromField: "start_date",
			ToField:   "end_date",
		},
		"Balance":     struct{ HasActiveYear bool }{result.HasActiveYear},
		"AppDataJSON": template.JS(raw),
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
		h.logger.Error(r.Context(), "failed to approve vacation", err)
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
		h.logger.Error(r.Context(), "failed to change vacation status", err)
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
