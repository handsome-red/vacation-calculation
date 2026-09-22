package register_form

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Handle(ctx context.Context, _ Query) (*Result, error) {

	positions := user.AllPositions()

	result := make([]PositionOption, 0, len(positions))
	for _, p := range positions {
		result = append(result, PositionOption{
			Value: p.String(),
			Title: p.Title(),
		})
	}
	return &Result{Positions: result}, nil
}
