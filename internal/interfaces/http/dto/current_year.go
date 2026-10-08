package dto

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type CurrentYearBlock struct {
	Range     string
	Total     int
	Earned    float64
	Used      int
	Remaining float64
	Today     string
}

func ToCurrentYearBlock(info vacation.CurrentYearInfo, now time.Time) CurrentYearBlock {
	return CurrentYearBlock{
		Range:     info.Year.From.Format(dateFormat) + " - " + info.Year.To.AddDate(0, 0, -1).Format(dateFormat),
		Total:     info.Year.Base + info.Year.Irregular + info.Year.Seniority,
		Earned:    info.Earned,
		Used:      info.Used,
		Remaining: info.Remaining,
		Today:     now.UTC().Format(dateFormat),
	}
}
