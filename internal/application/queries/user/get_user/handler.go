package get_user

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
	userRepo ports.UserRepository
	logger   ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	logger ports.Logger,
) *Handler {
	return &Handler{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	if query.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Парсим ID
	userID, err := user.ParseUserID(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Ищем пользователя
	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		h.logger.Error(ctx, "failed to find user", "user_id", query.UserID, "error", err)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &Result{
		ID:         u.ID().String(),
		Email:      u.Email().String(),
		FirstName:  u.Name().FirstName(),
		LastName:   u.Name().LastName(),
		MiddleName: u.Name().MiddleName(),
		Department: u.Department().String(),
		IsActive:   u.IsActive(),
		CreatedAt:  u.CreatedAt().String(),
		UpdatedAt:  u.UpdatedAt().String(),
	}, nil
}
