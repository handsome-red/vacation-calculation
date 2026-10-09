package get_vacation_form

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	userID, err := user.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return &Result{UserID: userID.String()}, nil
}
