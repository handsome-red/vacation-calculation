package list_users

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

func (h *Handler) Handle(ctx context.Context, query ports.Query) (*Result, error) {
	// Дефолтные значения
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Size < 1 || query.Size > 100 {
		query.Size = 20
	}

	// Получаем активных пользователей
	users, total, err := h.userRepo.FindUsers(ctx, query)
	if err != nil {
		h.logger.Error(ctx, "failed to get users", "error", err)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Формируем результат
	result := &Result{
		Users: make([]UserListItem, 0, len(users)),
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}

	for _, u := range users {
		result.Users = append(result.Users, UserListItem{
			ID:              u.ID().String(),
			Status:          u.Status().String(),
			Email:           u.Email().String(),
			FirstName:       u.FirstName(),
			LastName:        u.LastName(),
			MiddleName:      u.MiddleName(),
			BirthDate:       u.BirthDate().String(),
			Position:        u.Position().Title(),
			HiredAt:         u.HiredAt().String(),
			DepartmentTitle: u.Department().Title,
			DistrictTitle:   u.District().Title(),
			WorkdayDuration: u.WorkdayDuration().Int(),
			IsInvalid:       u.IsInvalid(),
			TotalExperience: u.ExperienceLabel(),
		})
	}

	return result, nil
}
