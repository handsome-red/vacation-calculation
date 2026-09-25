package get_vacation_form

import "context"

type Handler struct {
}

func NewHandler(
) *Handler {
	return &Handler{
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Query) (*Result, error) {
	return nil, nil
}