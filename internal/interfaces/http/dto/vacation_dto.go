package dto

import (
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type YearStat struct {
	From      string
	To        string
	Base      int
	Irregular int
	Seniority int
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
	return YearStat{
		From:      ys.From.Format(dateFormat),
		To:        ys.To.Format(dateFormat),
		Base:      ys.Base,
		Seniority: ys.Seniority,
		Irregular: ys.Irregular,
	}
}
