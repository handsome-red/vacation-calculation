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

	// Валидируем Status
	status, err := user.NewStatus(cmd.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	// Валидируем дату рождения
	birthDate, err := user.NewBirthDate(cmd.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("invalid birth date: %w", err)
	}

	// Валидируем должность
	position, err := user.NewPosition(cmd.Position)
	if err != nil {
		return nil, fmt.Errorf("invalid position: %w", err)
	}

	// Валидируем дату приёма
	hiredAt, err := user.NewHiredDate(cmd.HiredAt)
	if err != nil {
		return nil, fmt.Errorf("invalid hired date: %w", err)
	}

	// Валидируем район/отдел
	district := user.NewDistrict(cmd.DistrictCode)

	// Валидируем продолжительность рабочего дня
	workdayDuration, err := user.NewWorkdayDuration(cmd.WorkdayDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid workday duration: %w", err)
	}

	isInvalid := cmd.IsInvalid

	// Валидируем пароль и хешируем
	hashedPassword, err := h.hasher.HashPassword(ctx, cmd.Password)
	if err != nil {
		h.logger.Error(ctx, "failed to hash password", "error", err)
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Создаем пароль
	password := user.NewPasswordFromHash(hashedPassword)

	department := user.Department{Code: cmd.DepartmentCode}

	// Генерация UUID (Value Object)
	userID := user.GenerateUserID()

	newUser, err := user.NewUser(
		userID,
		status,
		cmd.LastName,
		cmd.FirstName,
		cmd.MiddleName,
		birthDate,
		position,
		hiredAt,
		department,
		district,
		workdayDuration,
		email,
		isInvalid,
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

	return &Result{
		UserID:     newUser.ID().String(),
		Email:      newUser.Email().String(),
		FirstName:  newUser.FirstName(),
		LastName:   newUser.LastName(),
		MiddleName: newUser.MiddleName(),
		CreatedAt:  newUser.CreatedAt().String(),
	}, nil
}
