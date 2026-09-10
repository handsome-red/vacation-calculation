package get_active_users

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
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
	// Дефолтные значения
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Size < 1 || query.Size > 100 {
		query.Size = 20
	}

	// Получаем активных пользователей
	users, total, err := h.userRepo.FindActive(ctx, query.Page, query.Size)
	if err != nil {
		h.logger.Error(ctx, "failed to get active users", "error", err)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Формируем результат
	result := &Result{
		Users:      make([]UserListItem, 0, len(users)),
		TotalCount: total,
		Page:       query.Page,
		PageSize:   query.Size,
	}

	for _, u := range users {
		result.Users = append(result.Users, UserListItem{
			ID:         u.ID().String(),
			Email:      u.Email().String(),
			FullName:   u.Name().FullName(),
			Department: u.Department(),
		})
	}

	return result, nil
}
