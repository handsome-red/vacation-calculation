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
	// ExcludeVacationID — отпуск, который не учитывается при проверках и расчёте
	// квот (переносимый). Заполняется только сервером, не из запроса.
	ExcludeVacationID string `json:"-"`
}

type Result struct {
	VacationIDs []string
	Allocations []vacation.Allocation
	// Warnings — правила, которые обошёл администратор.
	Warnings []string
	// Notices — предупреждения для всех: дни года используются позже срока переноса.
	Notices []string
}

func (c Command) Validate() error {
	if len(c.Periods) == 0 {
		return fmt.Errorf("%w: не передан ни один период", vacation.ErrInvalidInput)
	}
	return nil
}
