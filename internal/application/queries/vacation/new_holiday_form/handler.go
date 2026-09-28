package new_holiday_form

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	holidayRepo vacation.HolidayRepository	
}

func NewHandler(holidayRepo vacation.HolidayRepository	)*Handler {
	return &Handler{
		holidayRepo: holidayRepo,
	}
}


func (h *Handler)Handle(ctx context.Context, _ Query) (*Result, error) {
	return nil, nil
}