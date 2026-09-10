package register_user

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
	userRepo ports.UserRepository
	hasher   ports.PasswordHasher
	logger   ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	hasher ports.PasswordHasher,
	logger ports.Logger,
) *Handler {
	return &Handler{
		userRepo: userRepo,
		hasher:   hasher,
		logger:   logger,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	// Валидируем email
	email, err := user.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Проверяем уникальность email
	exists, err := h.userRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}
	if exists {
		return nil, user.ErrEmailAlreadyExists
	}

	// Валидируем пароль и хешируем
	hashedPassword, err := h.hasher.HashPassword(ctx, cmd.Password)
	if err != nil {
		h.logger.Error(ctx, "failed to hash password", "error", err)
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Создаем пароль
	password := user.NewPasswordFromHash(hashedPassword)

	// Создаем name (Value Object)
	name, err := user.NewName(cmd.FirstName, cmd.LastName, cmd.MiddleName)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %w", err)
	}

	// Генерация UUID (Value Object)
	userID := user.GenerateUserID()

	newUser, err := user.NewUser(
		userID,
		email,
		name,
		cmd.Department,
		password,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Сохраняем пользователя
	if err := h.userRepo.Save(ctx, newUser); err != nil {
		h.logger.Error(ctx, "failed to save user", "error", err)
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	h.logger.Info(ctx, "user registered successfully",
		"user_id", newUser.ID().String(),
		"email", newUser.Email().String(),
	)

	return &Result{
		UserID:    newUser.ID().String(),
		Email:     newUser.Email().String(),
		FullName:  newUser.Name().FullName(),
		CreatedAt: newUser.CreatedAt().String(),
	}, nil
}
