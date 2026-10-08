package new_holiday

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	holidayRepo ports.HolidayRepository
}

func NewHandler(holidayRepo ports.HolidayRepository) *Handler {
	return &Handler{
		holidayRepo: holidayRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {

	date, err := vacation.ParseDate(cmd.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", cmd.Date, err)
	}

	holiday, err := vacation.NewHoliday(cmd.Name, date)
	if err != nil {
		return nil, fmt.Errorf("failed to create holiday %s: %w", cmd.Name, err)
	}

	if err := h.holidayRepo.Save(ctx, holiday); err != nil {
		return nil, fmt.Errorf("failed to create holiday %s: %w", cmd.Name, err)
	}

	return &Result{
		Name: cmd.Name,
		Date: cmd.Date,
	}, nil
}
