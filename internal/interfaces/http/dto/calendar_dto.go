package dto

import (
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
)

type CalendarResponse struct {
	Year  int           `json:"year"`
	Month int           `json:"month"`
	Days  []CalendarDay `json:"days"`
}

type CalendarDay struct {
	Date        string          `json:"date"` // "2026-11-04"
	Day         int             `json:"day"`  // 4
	IsWeekend   bool            `json:"isWeekend"`
	IsHoliday   bool            `json:"isHoliday"`
	HolidayName string          `json:"holidayName,omitempty"`
	Events      []CalendarEvent `json:"events,omitempty"`
}

type CalendarEvent struct {
	Kind    string `json:"kind"`              // "vacation" | "shift"
	SubKind string `json:"subKind,omitempty"` // "APPROVED" | "unpaid_long" | ...
	Title   string `json:"title"`
	From    string `json:"from"` // "2026-10-28"
	To      string `json:"to"`   // "2026-11-05"
}

func ToCalendarResponse(r *get_calendar.Result) CalendarResponse {
	days := make([]CalendarDay, 0, len(r.Days))
	for _, d := range r.Days {
		events := make([]CalendarEvent, 0, len(d.Events))
		for _, e := range d.Events {
			events = append(events, CalendarEvent{
				Kind:    string(e.Kind),
				SubKind: e.SubKind,
				Title:   e.Title,
				From:    e.From.ISO(),
				To:      e.To.ISO(),
			})
		}
		days = append(days, CalendarDay{
			Date:        d.Date.ISO(),
			Day:         d.Date.Day(),
			IsWeekend:   d.IsWeekend,
			IsHoliday:   d.IsHoliday,
			HolidayName: d.HolidayName,
			Events:      events,
		})
	}
	return CalendarResponse{
		Year:  r.Year,
		Month: r.Month,
		Days:  days,
	}
}
