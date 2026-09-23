package change_email

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

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	// 1. Находим пользователя
	userID, err := user.NewUserID(cmd.UserID)
	if err != nil {
		return nil, user.ErrIDRequired
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Создаем новый email VO
	newEmail, err := user.NewEmail(cmd.NewEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// 3. Проверяем уникальность нового email
	exists, err := h.userRepo.ExistsByEmail(ctx, newEmail.String())
	if err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}
	if exists {
		return nil, user.ErrEmailAlreadyExists
	}

	// 4. Меняем email (бизнес-логика в Domain)
	if err := u.ChangeEmail(newEmail); err != nil {
		return nil, err
	}

	// 5. Сохраняем
	if err := h.userRepo.Save(ctx, u); err != nil {
		h.logger.Error(ctx, "failed to update user", "error", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	h.logger.Info(ctx, "email changed successfully",
		"user_id", u.ID().String(),
		"new_email", u.Email().String(),
	)

	return &Result{
		UserID:    u.ID().UUID(),
		NewEmail:  u.Email().String(),
		UpdatedAt: u.UpdatedAt(),
	}, nil
}
