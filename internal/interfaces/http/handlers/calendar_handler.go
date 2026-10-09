package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/new_holiday"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/new_holiday_form"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/dto"
)

type CalendarHandler struct {
	getCalendarUseCase    *get_calendar.Handler
	newHolidayUseCase     *new_holiday.Handler
	newHolidayFormUseCase *new_holiday_form.Handler
	templates             *Templates
	logger                ports.Logger
}

func NewCalendarHandler(
	getCalendarUseCase *get_calendar.Handler,
	newHolidayUseCase *new_holiday.Handler,
	newHolidayFormUseCase *new_holiday_form.Handler,
	templates *Templates,
	logger ports.Logger,
) *CalendarHandler {
	return &CalendarHandler{
		getCalendarUseCase:    getCalendarUseCase,
		newHolidayUseCase:     newHolidayUseCase,
		newHolidayFormUseCase: newHolidayFormUseCase,
		templates:             templates,
		logger:                logger,
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

	result, err := h.getCalendarUseCase.Handle(r.Context(), get_calendar.Query{
		UserID: userID,
		Year:   year,
		Month:  month,
	})
	if err != nil {
		h.logger.Error(r.Context(), "get user calendar failed",
			"user_id", userID, "year", year, "month", month, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(dto.ToCalendarResponse(result)); err != nil {
		// лог, но ответ уже частично отправлен
		h.logger.Error(r.Context(), "encode calendar response", "error", err)
	}
}

func (h *CalendarHandler) GetHolidayForm(w http.ResponseWriter, r *http.Request) {
	// TODO Добавить отображение календаря
	if _, err := h.newHolidayFormUseCase.Handle(r.Context(), new_holiday_form.Query{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.templates.Render(w, "get_holiday_form.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *CalendarHandler) NewHoliday(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	name := r.FormValue("name")

	cmd := new_holiday.Command{
		Date: dateStr,
		Name: name,
	}

	result, err := h.newHolidayUseCase.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_ = result
}
