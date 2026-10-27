package login

import (
	"context"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Handler struct {
	userRepo ports.UserRepository
	hasher   ports.PasswordHasher
}

func NewHandler(
	userRepo ports.UserRepository,
	hasher ports.PasswordHasher,
) *Handler {
	return &Handler{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	email, err := user.NewEmail(cmd.Email)
	if err != nil || cmd.Password == "" {
		return nil, ErrInvalidCredentials
	}
	u, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: find by email: %w", err)
	}

	if !u.Status().IsActive() {
		return nil, ErrInvalidCredentials
	}
	ok, err := h.hasher.Verify(ctx, u.Password().String(), cmd.Password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !ok {
		return nil, ErrInvalidCredentials
	}

	return &Result{
		UserID:  u.ID().String(),
		Email:   u.Email().Value(),
		IsAdmin: u.IsAdmin(),
	}, nil
}
