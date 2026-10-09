package login

import (
	"context"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Handler struct {
	userRepo ports.AuthRepository
	hasher   ports.PasswordHasher
}

func NewHandler(
	userRepo ports.AuthRepository,
	hasher ports.PasswordHasher,
) *Handler {
	return &Handler{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	user, err := h.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: find by email: %w", err)
	}

	ok, err := h.hasher.Verify(ctx, user.PasswordHash().Value(), cmd.Password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !ok {
		return nil, ErrInvalidCredentials
	}

	return &Result{
		UserID: user.ID().String(),
		Email:  user.Email().Value(),
	}, nil
}
