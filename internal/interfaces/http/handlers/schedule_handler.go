package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/schedule/approve_schedule"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/schedule/get_schedule"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

// ScheduleHandler — график отпусков для кадровика: проверка, утверждение, выгрузка в Word.
type ScheduleHandler struct {
	get       *get_schedule.Handler
	approve   *approve_schedule.Handler
	renderer  ports.ScheduleRenderer
	templates *Templates
	logger    ports.Logger
}

func NewScheduleHandler(
	get *get_schedule.Handler,
	approve *approve_schedule.Handler,
	renderer ports.ScheduleRenderer,
	templates *Templates,
	logger ports.Logger,
) *ScheduleHandler {
	return &ScheduleHandler{get: get, approve: approve, renderer: renderer, templates: templates, logger: logger}
}

// ScheduleYear — год графика по умолчанию: с октября формируется график на следующий год.
func ScheduleYear(now time.Time) int {
	if now.Month() >= time.October {
		return now.Year() + 1
	}
	return now.Year()
}

func scheduleQuery(r *http.Request) get_schedule.Query {
	year := ScheduleYear(time.Now())
	if y, err := strconv.Atoi(r.FormValue("year")); err == nil && y >= 2000 && y <= 2100 {
		year = y
	}
	return get_schedule.Query{Year: year, DistrictCode: r.FormValue("district")}
}

func (h *ScheduleHandler) Page(w http.ResponseWriter, r *http.Request) {
	q := scheduleQuery(r)
	res, err := h.get.Handle(r.Context(), q)
	if err != nil {
		h.logger.Error(r.Context(), "schedule page", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Schedule": res,
		"PrevYear": q.Year - 1,
		"NextYear": q.Year + 1,
		"Approved": r.URL.Query().Get("approved"),
	}
	if err := h.templates.Render(w, "schedule.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ScheduleHandler) Approve(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	q := scheduleQuery(r)
	res, err := h.approve.Handle(r.Context(), approve_schedule.Command{Year: q.Year, DistrictCode: q.DistrictCode})
	if err != nil {
		h.logger.Error(r.Context(), "approve schedule", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Info(r.Context(), "schedule approved", "year", q.Year, "district", q.DistrictCode, "vacations", res.Approved)
	http.Redirect(w, r, fmt.Sprintf("/admin/schedule?year=%d&district=%s&approved=%d", q.Year, url.QueryEscape(q.DistrictCode), res.Approved), http.StatusSeeOther)
}

// Report — график отпусков на год в Word.
func (h *ScheduleHandler) Report(w http.ResponseWriter, r *http.Request) {
	q := scheduleQuery(r)
	report, err := h.get.Report(r.Context(), q)
	if err != nil {
		h.logger.Error(r.Context(), "schedule report", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doc, err := h.renderer.RenderSchedule(*report)
	if err != nil {
		h.logger.Error(r.Context(), "render schedule report", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := fmt.Sprintf("График отпусков %d", q.Year)
	ascii := fmt.Sprintf("grafik-otpuskov-%d", q.Year)
	if q.DistrictCode != "" {
		name += " " + q.DistrictCode
		ascii += "-" + q.DistrictCode
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.docx"; filename*=UTF-8''%s`, ascii, url.PathEscape(name+".docx")))
	_, _ = w.Write(doc)
}
