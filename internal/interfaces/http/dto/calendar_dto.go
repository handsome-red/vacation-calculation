package dto

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_calendar"
)

type CalendarResponse struct {
	Year   int        `json:"year"`
	Month  int        `json:"month"`
	Months []MonthDTO `json:"months"`
}

type UserCalendarResponse struct {
	Year  int        `json:"year"`
	Month int        `json:"month"`
	Days  []UDayDTO  `json:"days"`
}

type MonthDTO struct {
	Month int           `json:"month"` // 1..12
	Days  []CalendarDay `json:"days"`
}

type CalendarDay struct {
	Date        string `json:"date"`               // "2026-11-04"
	Day         int    `json:"day"`                // 4
	IsWeekend   bool   `json:"isWeekend"`
	IsHoliday   bool   `json:"isHoliday"`
	HolidayName string `json:"holidayName,omitempty"`
}

type UDayDTO struct {
	Date             string `json:"date"`
	IsWeekend        bool   `json:"isWeekend"`
	IsHoliday        bool   `json:"isHoliday"`
	HolidayName      string `json:"holidayName,omitempty"`
	IsWorkSaturday   bool   `json:"isWorkSaturday"`
	WorkSaturdayName string `json:"workSaturdayName,omitempty"`
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
			Date:             d.Date.ISO(),
			IsWeekend:        d.IsWeekend,
			IsHoliday:        d.IsHoliday,
			HolidayName:      d.HolidayName,
			IsWorkSaturday:   d.IsWorkSaturday,
			WorkSaturdayName: d.WorkSaturdayName,
		}
	}

	return UserCalendarResponse{
		Year:  r.Year,
		Month: r.Month,
		Days:  days,
	}
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
