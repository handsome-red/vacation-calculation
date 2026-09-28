package get_calendar

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	holidayRepo vacation.HolidayRepository
}

func NewHandler(holidayRepo vacation.HolidayRepository) *Handler {
	return &Handler{
		holidayRepo: holidayRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	holidays, err := h.holidayRepo.ListByYear(ctx, query.Year)
	if err != nil {
		return nil, fmt.Errorf("list holidays: %w", err)
	}

	items := make([]holidayItem, 0, len(holidays))
	for _, h := range holidays {
		items = append(items, holidayItem{
			Name: h.Name(),
			Date: h.Date().String(),
		})
	} 

	raw, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marhal holidays: %w", err)
	}

	return &Result{
		Year:         query.Year,
		HolidaysJSON: string(raw),
	}, nil
}