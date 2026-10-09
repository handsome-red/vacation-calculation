package get_current_user

import "context"

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	return &Result{
		User: auth.AuthUser{},
	}, nil
}
