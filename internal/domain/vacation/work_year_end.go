package vacation

import (
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// WorkYear — сохранённый рабочий год с номером Number (с единицы): завершившийся или с
// границами, заданными администратором. Его границы больше не пересчитываются.
type WorkYear struct {
	Number int
	Start  date.Date
	End    date.Date // последний день, включительно
	Manual bool
}

// WorkYearEnd — последний день рабочего года с номером Number, заданный администратором.
// Заменяет расчёт длительности года; следующий год начинается на другой день.
type WorkYearEnd struct {
	Number  int
	LastDay date.Date
}

// VacationWorkYear — новый рабочий год отпуска после изменения границ лет.
type VacationWorkYear struct {
	ID    VacationID
	Start date.Date
}

// ValidateWorkYears проверяет сохранённые годы: идут подряд с первого, первый начинается
// в день приёма, каждый следующий — на другой день после конца предыдущего.
func ValidateWorkYears(hiredAt time.Time, years []WorkYear) error {
	next := DateOf(hiredAt)
	for i, y := range years {
		switch {
		case y.Number != i+1:
			return fmt.Errorf("%w: ожидался год %d, сохранён %d", ErrWorkYearsInvalid, i+1, y.Number)
		case y.Start != next:
			return fmt.Errorf("%w: год %d начинается %s, ожидалось %s", ErrWorkYearsInvalid, y.Number, y.Start.ISO(), next.ISO())
		case y.End.Before(y.Start):
			return fmt.Errorf("%w: год %d кончается раньше начала", ErrWorkYearsInvalid, y.Number)
		}
		next = DateOf(y.End.Time().AddDate(0, 0, 1))
	}
	return nil
}
