package dto

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type CurrentYearBlock struct {
	Range     string
	Base      int
	Irregular int
	Seniority int
	Earned    float64
	Remaining float64
	Today     string
}

func ToCurrentYearBlock(info vacation.CurrentYearInfo, now time.Time) CurrentYearBlock {
	block := CurrentYearBlock{
		Today: now.UTC().Format(dateFormat),
	}

	block.Range = info.Year.From.Format(dateFormat) + " — " + info.Year.To.Format(dateFormat)
	block.Base = info.Year.Base
	block.Irregular = info.Year.Irregular
	block.Seniority = info.Year.Seniority
	block.Earned = info.Earned
	block.Remaining = info.Remaining
	return block
}
