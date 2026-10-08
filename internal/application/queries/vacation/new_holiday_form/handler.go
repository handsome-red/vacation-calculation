package new_holiday_form

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type Handler struct {
	holidayRepo ports.HolidayRepository
}

func NewHandler(holidayRepo ports.HolidayRepository) *Handler {
	return &Handler{
		holidayRepo: holidayRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, _ Query) (*Result, error) {
	return nil, nil
}
