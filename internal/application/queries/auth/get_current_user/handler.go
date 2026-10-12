package get_current_user

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
)

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
