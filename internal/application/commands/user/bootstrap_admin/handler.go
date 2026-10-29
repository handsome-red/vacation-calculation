package bootstrap_admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

var ErrPasswordRequired = errors.New("bootstrap admin: ADMIN_PASSWORD is required to create a new admin")

type Handler struct {
	userRepo ports.UserRepository
	register *register_user.Handler
}

func NewHandler(userRepo ports.UserRepository, register *register_user.Handler) *Handler {
	return &Handler{userRepo: userRepo, register: register}
}

// Handle создаёт первого администратора, если в системе ещё нет ни одного.
// Если пользователь с таким email уже существует, ему выдаются права администратора.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	exists, err := h.userRepo.ExistsAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return &Result{}, nil
	}

	email, err := user.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("bootstrap admin: invalid ADMIN_EMAIL: %w", err)
	}

	result := &Result{}
	u, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, fmt.Errorf("bootstrap admin: find user: %w", err)
	}
	if u == nil {
		if cmd.Password == "" {
			return nil, ErrPasswordRequired
		}
		created, err := h.register.Handle(ctx, register_user.Command{
			Email:           cmd.Email,
			Password:        cmd.Password,
			LastName:        cmd.LastName,
			FirstName:       cmd.FirstName,
			MiddleName:      cmd.MiddleName,
			BirthDate:       cmd.BirthDate,
			Position:        cmd.Position,
			HiredAt:         cmd.HiredAt,
			DistrictCode:    cmd.DistrictCode,
			DepartmentCode:  cmd.DepartmentCode,
			WorkdayDuration: 8,
		})
		if err != nil {
			return nil, fmt.Errorf("bootstrap admin: register: %w", err)
		}
		id, err := user.ParseUserID(created.UserID)
		if err != nil {
			return nil, fmt.Errorf("bootstrap admin: %w", err)
		}
		if u, err = h.userRepo.FindByID(ctx, id); err != nil {
			return nil, fmt.Errorf("bootstrap admin: load created user: %w", err)
		}
		result.Created = true
	} else {
		result.Promoted = true
	}

	u.GrantAdmin()
	if err := h.userRepo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("bootstrap admin: save: %w", err)
	}

	result.UserID = u.ID().String()
	return result, nil
}
