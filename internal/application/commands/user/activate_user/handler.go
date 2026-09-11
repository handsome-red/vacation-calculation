package activate_user

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

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// 1. Находим пользователя
	userID, err := user.NewUserID(cmd.UserID)
	if err != nil {
		return err
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 2. Активируем (бизнес-логика в Domain)
	if err := u.Activate(); err != nil {
		return err
	}

	// 3. Сохраняем
	if err := h.userRepo.Save(ctx, u); err != nil {
		h.logger.Error(ctx, "failed to update user", "error", err)
		return fmt.Errorf("failed to activate user: %w", err)
	}

	h.logger.Info(ctx, "user activated successfully",
		"user_id", u.ID().String(),
	)

	return nil
}
