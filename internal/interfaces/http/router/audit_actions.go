package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	authmiddleware "github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

// auditActions — изменяющие маршруты, которые попадают в журнал изменений. Новый
// изменяющий маршрут нужно добавить и сюда.
func auditActions(vacations ports.VacationRepository) []authmiddleware.AuditAction {
	// describeVacation — отпуск из маршрута после изменения: вид, даты и статус.
	describeVacation := func(r *http.Request) string {
		userID, err := user.ParseUserID(chi.URLParam(r, "userId"))
		if err != nil {
			return ""
		}
		id, err := vacation.ParseVacationID(chi.URLParam(r, "id"))
		if err != nil {
			return ""
		}
		v, err := vacations.FindByID(r.Context(), userID, id)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%s, %s, статус: %s", v.Kind().Label(), period(v.StartDate(), v.EndDate()), v.Status().String())
	}
	form := func(format string, fields ...string) func(r *http.Request) string {
		return func(r *http.Request) string {
			values := make([]any, len(fields))
			for i, f := range fields {
				values[i] = r.FormValue(f)
			}
			return fmt.Sprintf(format, values...)
		}
	}

	return []authmiddleware.AuditAction{
		{Method: http.MethodPost, Pattern: "/user/register", Action: "Добавлен сотрудник"},
		{Method: http.MethodPost, Pattern: "/users/{id}/edit", Action: "Изменены данные сотрудника", TargetParam: "id",
			Describe: func(r *http.Request) string {
				if r.FormValue("password") != "" {
					return "пароль изменён"
				}
				return ""
			}},
		{Method: http.MethodPost, Pattern: "/users/{id}/work-years/{number}", Action: "Изменены границы рабочего года", TargetParam: "id",
			Describe: func(r *http.Request) string {
				if r.FormValue("action") == "reset" {
					return fmt.Sprintf("год %s: возвращён расчёт", chi.URLParam(r, "number"))
				}
				return fmt.Sprintf("год %s: %s", chi.URLParam(r, "number"), isoPeriod(r.FormValue("start"), r.FormValue("end")))
			}},
		{Method: http.MethodDelete, Pattern: "/api/v1/users/{id}", Action: "Сотрудник деактивирован", TargetParam: "id"},
		{Method: http.MethodPost, Pattern: "/api/v1/users/{userId}/vacations", Action: "Оформлен отпуск", TargetParam: "userId"},
		{Method: http.MethodPost, Pattern: "/api/v1/users/{userId}/vacations/{id}/approve", Action: "Отпуск утверждён", TargetParam: "userId", Describe: describeVacation},
		{Method: http.MethodPost, Pattern: "/api/v1/users/{userId}/vacations/{id}/status", Action: "Изменён статус отпуска", TargetParam: "userId", Describe: describeVacation},
		{Method: http.MethodPost, Pattern: "/api/v1/users/{userId}/vacations/{id}/delete", Action: "Отпуск удалён из базы", TargetParam: "userId"},
		{Method: http.MethodPost, Pattern: "/api/v1/users/{userId}/vacations/{id}/transfer", Action: "Отпуск перенесён", TargetParam: "userId",
			Describe: func(r *http.Request) string {
				s := "на " + isoPeriod(r.FormValue("new_start"), r.FormValue("new_end"))
				if reason := strings.TrimSpace(r.FormValue("reason")); reason != "" {
					s += ", причина: " + reason
				}
				return s
			}},
		{Method: http.MethodPost, Pattern: "/api/v1/users/{userId}/shift", Action: "Добавлен сдвиг рабочего года", TargetParam: "userId",
			Describe: func(r *http.Request) string {
				return isoPeriod(r.FormValue("date_from"), r.FormValue("date_to"))
			}},
		{Method: http.MethodPost, Pattern: "/admin/holidays", Action: "Изменён производственный календарь", Describe: form("%s: %s %s", "date", "day_type", "name")},
		{Method: http.MethodPost, Pattern: "/admin/holidays/delete", Action: "Удалён день производственного календаря", Describe: form("%s", "date")},
		{Method: http.MethodPost, Pattern: "/admin/holidays/fill", Action: "Заполнены праздничные дни", Describe: form("год %s", "year")},
		{Method: http.MethodPost, Pattern: "/admin/schedule/approve", Action: "Утверждён график отпусков",
			Describe: func(r *http.Request) string {
				s := "год " + r.FormValue("year")
				if d := r.FormValue("department"); d != "" {
					s += ", отдел " + d
				}
				return s
			}},
	}
}

// period — даты отпуска в виде «01.07.2027–14.07.2027».
func period(start, end date.Date) string {
	return start.Time().Format("02.01.2006") + "–" + end.Time().Format("02.01.2006")
}

// isoPeriod — то же для дат из формы (ГГГГ-ММ-ДД); непонятные даты выводятся как есть.
func isoPeriod(start, end string) string {
	format := func(s string) string {
		if d, err := date.ParseDate(s); err == nil {
			return d.Time().Format("02.01.2006")
		}
		return s
	}
	return format(start) + "–" + format(end)
}
