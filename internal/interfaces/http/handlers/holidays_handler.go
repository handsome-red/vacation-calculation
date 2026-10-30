package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/calendar/new_exception"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/new_calendar_day_form"
	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
)

// Страница администратора: особые дни производственного календаря за год.

type calendarDayFormData struct {
	Year     int
	Days     []new_calendar_day_form.Day
	Form     new_calendar_day.Command
	Errors   []string
	Notice   string
	PrevYear int
	NextYear int
}

func holidaysYear(r *http.Request) int {
	if y, err := strconv.Atoi(r.FormValue("year")); err == nil && y >= 2000 && y <= 2100 {
		return y
	}
	return time.Now().Year()
}

func (h *CalendarHandler) renderHolidays(w http.ResponseWriter, r *http.Request, status int, data calendarDayFormData) {
	list, err := h.newCalendarDayFormUseCase.Handle(r.Context(), new_calendar_day_form.Query{Year: data.Year})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Days = list.Days
	data.PrevYear, data.NextYear = data.Year-1, data.Year+1
	if data.Form.Type == "" {
		data.Form.Type = string(calendar.DayTypeHoliday)
	}
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	if err := h.templates.Render(w, "get_calendar_day_form.html", data); err != nil {
		h.logger.Error(r.Context(), "render holidays page", "error", err)
	}
}

func (h *CalendarHandler) GetCalendarDayForm(w http.ResponseWriter, r *http.Request) {
	data := calendarDayFormData{Year: holidaysYear(r)}
	switch r.URL.Query().Get("done") {
	case "saved":
		data.Notice = "День сохранён."
	case "deleted":
		data.Notice = "День удалён."
	case "filled":
		data.Notice = "Добавлено постоянных праздников: " + r.URL.Query().Get("n") + "."
	}
	h.renderHolidays(w, r, http.StatusOK, data)
}

// NewCalendarDay добавляет или изменяет (по дате) особый день календаря.
func (h *CalendarHandler) NewCalendarDay(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	cmd := new_calendar_day.Command{
		Date: r.FormValue("date"),
		Type: r.FormValue("day_type"),
		Name: r.FormValue("name"),
	}

	result, err := h.newCalendarDayUseCase.Handle(r.Context(), cmd)
	if err != nil {
		h.renderHolidays(w, r, http.StatusBadRequest, calendarDayFormData{Year: holidaysYear(r), Form: cmd, Errors: []string{err.Error()}})
		return
	}

	http.Redirect(w, r, "/admin/holidays?done=saved&year="+result.Date[:4], http.StatusSeeOther)
}

func (h *CalendarHandler) DeleteCalendarDay(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	if err := h.newCalendarDayUseCase.Delete(r.Context(), r.FormValue("date")); err != nil {
		h.renderHolidays(w, r, http.StatusBadRequest, calendarDayFormData{Year: holidaysYear(r), Errors: []string{err.Error()}})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/holidays?done=deleted&year=%d", holidaysYear(r)), http.StatusSeeOther)
}

// FillHolidays добавляет постоянные праздники выбранного года.
func (h *CalendarHandler) FillHolidays(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	year := holidaysYear(r)
	n, err := h.newCalendarDayUseCase.FillFixedHolidays(r.Context(), year)
	if err != nil {
		h.renderHolidays(w, r, http.StatusInternalServerError, calendarDayFormData{Year: year, Errors: []string{err.Error()}})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/holidays?done=filled&n=%d&year=%d", n, year), http.StatusSeeOther)
}
