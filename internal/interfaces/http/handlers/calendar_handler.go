package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/new_calendar_day"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/new_calendar_day_form"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/dto"
)

type CalendarHandler struct {
	getUserCalendarUseCase    *get_user_calendar.Handler
	getCalendarUseCase        *get_calendar.Handler
	newCalendarDayUseCase     *new_calendar_day.Handler
	newCalendarDayFormUseCase *new_calendar_day_form.Handler
	templates                 *Templates
	logger                    ports.Logger
}

func NewCalendarHandler(
	getCalendarUseCase *get_calendar.Handler,
	newCalendarDayUseCase *new_calendar_day.Handler,
	newCalendarDayFormUseCase *new_calendar_day_form.Handler,
	templates *Templates,
	logger ports.Logger,
) *CalendarHandler {
	return &CalendarHandler{
		getCalendarUseCase:        getCalendarUseCase,
		newCalendarDayUseCase:     newCalendarDayUseCase,
		newCalendarDayFormUseCase: newCalendarDayFormUseCase,
		templates:                 templates,
		logger:                    logger,
	}
}

func (h *CalendarHandler) GetUserCalendar(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	year := time.Now().Year()
	if s := r.URL.Query().Get("year"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "invalid year", http.StatusBadRequest)
			return
		}
		year = v
	}

	month := int(time.Now().Month())
	if s := r.URL.Query().Get("month"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 || v > 12 {
			http.Error(w, "invalid month", http.StatusBadRequest)
			return
		}
		month = v
	}

	result, err := h.getUserCalendarUseCase.Handle(r.Context(), get_user_calendar.Query{
		UserID: userID,
		Year:   year,
	})
	if err != nil {
		h.logger.Error(r.Context(), "get user calendar failed",
			"user_id", userID, "year", year, "month", month, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(dto.ToCalendarResponse(result)); err != nil {
		h.logger.Error(r.Context(), "encode calendar response", "error", err)
		return
	}
}

func (h *CalendarHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	if s := r.URL.Query().Get("year"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "invalid year", http.StatusBadRequest)
			return
		}
		year = v
	}

	month := int(time.Now().Month())
	if s := r.URL.Query().Get("month"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 || v > 12 {
			http.Error(w, "invalid month", http.StatusBadRequest)
			return
		}
		month = v
	}

	result, err := h.getCalendarUseCase.Handle(r.Context(), get_calendar.Query{
		Year: year,
	})
	if err != nil {
		h.logger.Error(r.Context(), "get user calendar failed",
			"year", year, "month", month, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(dto.ToCalendarResponse(result)); err != nil {
		h.logger.Error(r.Context(), "encode calendar response", "error", err)
		return
	}
}

func (h *CalendarHandler) GetCalendarDayForm(w http.ResponseWriter, r *http.Request) {
	// TODO Добавить отображение календаря
	if _, err := h.newCalendarDayFormUseCase.Handle(r.Context(), new_calendar_day_form.Query{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.templates.Render(w, "get_calendar_day_form.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *CalendarHandler) NewCalendarDay(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	name := r.FormValue("name")

	cmd := new_calendar_day.Command{
		Date: dateStr,
		Name: name,
	}

	result, err := h.newCalendarDayUseCase.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_ = result
}
