package create_vacations

import (
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Period struct {
	Kind  vacation.Kind
	Start date.Date
	End   date.Date
}

type Command struct {
	UserID  string   `json:"userId"`
	Periods []Period `json:"periods"`
}

type Result struct {
	VacationID string
	Error      error
}

func (c Command) Validate() error {
	if len(c.Periods) == 0 {
		return fmt.Errorf("%w: не передан ни один период", vacation.ErrInvalidInput)
	}
	return nil
}
