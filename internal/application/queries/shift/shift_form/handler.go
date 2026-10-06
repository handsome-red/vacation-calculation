package shift_form

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Handle(ctx context.Context, _ Query) (*Result, error) {
	kinds := vacation.AllShiftKindName()
	result := make([]ShiftKindOption, 0, len(kinds))
	for _, k := range kinds {
		result = append(result, ShiftKindOption{
			Code:  string(k),
			Title: k.String(),
		})
	}

	return &Result{
		Kinds: result,
	}, nil
}
