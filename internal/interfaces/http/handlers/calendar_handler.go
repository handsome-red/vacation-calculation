package handlers
import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/new_holiday"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/new_holiday_form"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type CalendarHandler struct {
	getCalendarUseCase *get_calendar.Handler
	newHolidayUseCase *new_holiday.Handler
	newHolidayFormUseCase *new_holiday_form.Handler
	templates *Templates
	logger ports.Logger
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
		newHolidayUseCase: newHolidayUseCase,
		newHolidayFormUseCase: newHolidayFormUseCase,
		templates:             templates,
		logger: logger,
	}
}

func (h *CalendarHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	if s := r.URL.Query().Get("year"); s != "" {
		y, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		year = y
	}

	result, err := h.getCalendarUseCase.Handle(r.Context(), get_calendar.Query{Year: year})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Year":         year,
		"PrevYear":     year - 1,
		"NextYear":     year + 1,
		"HolidaysJSON": template.JS(result.HolidaysJSON),
	}
	if err := h.templates.Render(w, "calendar.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	h.logger.Info(r.Context(), "create new holiday", result)
}