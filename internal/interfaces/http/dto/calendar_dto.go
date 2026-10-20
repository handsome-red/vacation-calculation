package dto

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type CalendarResponse struct {
	Year   int        `json:"year"`
	Month  int        `json:"month"`
	Months []MonthDTO `json:"months"`
}

type UserCalendarResponse struct {
	Year  int       `json:"year"`
	Month int       `json:"month"`
	Days  []UDayDTO `json:"days"`
}

type MonthDTO struct {
	Month int           `json:"month"` // 1..12
	Days  []CalendarDay `json:"days"`
}

type CalendarEvent struct {
	Kind    string `json:"kind"` // "vacation" | "shift"
	SubKind string `json:"subKind,omitempty"`
	Title   string `json:"title"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type CalendarDay struct {
	Date        string          `json:"date"` // "2026-11-04"
	Day         int             `json:"day"`  // 4
	IsWeekend   bool            `json:"isWeekend"`
	IsHoliday   bool            `json:"isHoliday"`
	HolidayName string          `json:"holidayName,omitempty"`
	Events      []CalendarEvent `json:"events,omitempty"`
}

type UEventDTO struct {
	Kind    string `json:"kind"` // "vacation" | "shift"
	SubKind string `json:"subKind,omitempty"`
	Title   string `json:"title"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type UDayDTO struct {
	Date        string      `json:"date"`
	IsWeekend   bool        `json:"isWeekend"`
	IsHoliday   bool        `json:"isHoliday"`
	HolidayName string      `json:"holidayName,omitempty"`
	Events      []UEventDTO `json:"events,omitempty"`
}

func ToCalendarResponse(r *get_calendar.Result) CalendarResponse {
	cal := r.Calendar

	months := make([]MonthDTO, 0, len(cal.Month))
	for _, m := range cal.Month {
		days := make([]CalendarDay, 0, len(m.Days))
		for _, d := range m.Days {
			days = append(days, CalendarDay{
				Date:        d.Date.ISO(),
				Day:         d.Date.Day(),
				IsWeekend:   d.Kind.IsWeekend(),
				IsHoliday:   d.Kind.IsHoliday(),
				HolidayName: d.Name,
			})
		}

		months = append(months, MonthDTO{
			Month: int(m.Number),
			Days:  days,
		})
	}

	return CalendarResponse{
		Year:   cal.Year,
		Months: months,
	}
}

func ToUserCalendarResponse(r *get_user_calendar.Result) UserCalendarResponse {
	days := make([]UDayDTO, len(r.Days))
	for i, d := range r.Days {
		days[i] = UDayDTO{
			Date:        d.Date.ISO(),
			IsWeekend:   d.IsWeekend,
			IsHoliday:   d.IsHoliday,
			HolidayName: d.HolidayName,
			// IsWorkSaturday:   d.IsWorkSaturday,
			// WorkSaturdayName: d.WorkSaturdayName,
			Events: eventsForDay(d.Date, r.Vacations),
		}
	}

	return UserCalendarResponse{
		Year:  r.Year,
		Month: r.Month,
		Days:  days,
	}
}

func eventsForDay(date date.Date, vacation []*vacation.Vacation) []UEventDTO {
	result := make([]UEventDTO, 0, 1)

	for _, v := range vacation {
		if !date.Before(v.StartDate()) && !date.After(v.EndDate()) {
			result = append(result, UEventDTO{
				Kind:    "vacation",
				SubKind: v.Status().String(),
				Title:   "отпуск",
				From:    v.StartDate().ISO(),
				To:      v.EndDate().ISO(),
			})
		}
	}
	return result
}

// ToMonthDays converts user calendar Result into a list of days for rendering as month-based grid
func ToMonthDays(year int, month time.Month, r *get_user_calendar.Result) []CalendarDay {
	days := make([]CalendarDay, 0, 31)
	for _, d := range r.Days {
		isMonthDay := d.Date.Month() == month && d.Date.Year() == year
		if !isMonthDay {
			continue
		}
		days = append(days, CalendarDay{
			Date:        d.Date.ISO(),
			Day:         d.Date.Day(),
			IsWeekend:   d.IsWeekend,
			IsHoliday:   d.IsHoliday,
			HolidayName: d.HolidayName,
		})
	}
	return days
}
