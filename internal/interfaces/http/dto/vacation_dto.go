package dto

import (
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type CurrentYear struct {
	From string
	To   string
}

type ShiftDTO struct {
	From   string
	To     string
	Reason string
}

type YearStat struct {
	From      string
	To        string
	Base      int
	Irregular int
	Seniority int
	Shifts    []ShiftDTO
}

type VacationBlock struct {
	Stats []YearStat
}

var dateFormat = "02.01.2006"

func ToYearStats(stats []vacation.YearStat) []YearStat {
	out := make([]YearStat, 0, len(stats))
	for _, ys := range stats {
		out = append(out, ToYearStat(ys))
	}
	return out
}

func ToYearStat(ys vacation.YearStat) YearStat {
	shifts := make([]ShiftDTO, 0, len(ys.Shifts))
	for _, s := range ys.Shifts {
		shifts = append(shifts, ShiftDTO{
			From:   s.From.Format(dateFormat),
			To:     s.To.Format(dateFormat),
			Reason: s.Kind.String(),
		})
	}
	return YearStat{
		From:      ys.From.Format(dateFormat),
		To:        ys.To.Format(dateFormat),
		Base:      ys.Base,
		Seniority: ys.Seniority,
		Irregular: ys.Irregular,
		Shifts:    shifts,
	}
}

func ToCurrentYear(stat vacation.YearStat) CurrentYear {
	return CurrentYear{
		From: stat.From.Format(dateFormat),
		To:   stat.To.Format(dateFormat),
	}
}
