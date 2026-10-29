package create_vacations

import (
	"encoding/json"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type periodJSON struct {
	Kind      string `json:"kind"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

func (p *Period) UnmarshalJSON(data []byte) error {
	var raw periodJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: период: %v", vacation.ErrInvalidInput, err)
	}

	// Вид отпуска определяет распределение по квотам; поле оставлено для совместимости.
	if raw.Kind == "" {
		raw.Kind = string(vacation.KindBase)
	}
	kind, err := vacation.NewKind(raw.Kind)
	if err != nil {
		return fmt.Errorf("%w: неизвестный тип периода %q", vacation.ErrInvalidInput, raw.Kind)
	}

	start, err := date.ParseDate(raw.StartDate)
	if err != nil {
		return fmt.Errorf("%w: не удалось разобрать дату начала %q", vacation.ErrInvalidInput, raw.StartDate)
	}
	end, err := date.ParseDate(raw.EndDate)
	if err != nil {
		return fmt.Errorf("%w: не удалось разобрать дату окончания %q", vacation.ErrInvalidInput, raw.EndDate)
	}

	*p = Period{Kind: kind, Start: start, End: end}
	return nil
}
