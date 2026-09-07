package userservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

const minPassLength = 7 // минимальная длина пароля

type RegisterUserCommand struct {
	Email      string
	Password   string
	FirstName  string
	LastName   string
	MiddleName string
	Department string
}

// UpdateUserEmailCommand - команда для изменения email.
type UpdateUserEmailCommand struct {
	UserID   uuid.UUID
	NewEmail string
}

// DeactivateUserCommand - команда для деактивации пользователя.
type ChangeUserStatusCommand struct {
	UserID uuid.UUID
}

type ChangeEmailResult struct {
	UserID    uuid.UUID `json:"user_id"`
	NewEmail  string    `json:"new_email"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDTO - объект для передачи данных о пользователе клиенту.
type UserDTO struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`
}

type UserApplicationService interface {
	RegisterUser(ctx context.Context, cmd RegisterUserCommand) (*UserDTO, error)
	GetUserByID(ctx context.Context, userID string) (*UserDTO, error)
	ChangeEmail(ctx context.Context, cmd UpdateUserEmailCommand) (*ChangeEmailResult, error)
	DeactivateUser(ctx context.Context, cmd ChangeUserStatusCommand) error
	ActivateUser(ctx context.Context, cmd ChangeUserStatusCommand) error
	GetActiveUsers(ctx context.Context) ([]*UserDTO, error)
}

type userApplicationService struct {
	userRepo user.UserRepository
	// userService user.UserServise
}

func NewUserApplicationService(
	userRepo user.UserRepository,
	// userService domain.UserService,
) UserApplicationService {
	return &userApplicationService{
		userRepo: userRepo,
		// userService: userService,
	}
}

func (s *userApplicationService) RegisterUser(
	ctx context.Context,
	cmd RegisterUserCommand,
) (*UserDTO, error) {
	if cmd.Email == "" {
		return nil, errors.New("email is required")
	}

	if cmd.Password == "" {
		return nil, errors.New("password is required")
	}

	if cmd.FirstName == "" || cmd.LastName == "" || cmd.MiddleName == "" {
		return nil, errors.New("first name and last name are required")
	}

	if len(cmd.Password) < minPassLength {
		return nil, fmt.Errorf("password must be at least %d characters long.", minPassLength)
	}

	if cmd.Department == "" {
		return nil, errors.New("deparment is required")
	}

	email, err := user.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	hashedPassword := hashPassword(cmd.Password)
	// password :=
	userUUID := uuid.New()
	userID := user.NewUserID(userUUID)

	user, err := user.NewUser(
		userID,
		email,
		hashedPassword,
		cmd.FirstName,
		cmd.LastName,
		cmd.MiddleName,
		cmd.Department,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// unique, err := s.userService.IsUnique(email.Value())
	return s.toUserDTO(user), nil
}

func (s *userApplicationService) GetUserByID(
	ctx context.Context,
	userID string,
) (*UserDTO, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	id, err := user.NewUserID()
}

func (s *userApplicationService) ChangeEmail(ctx context.Context, cmd UpdateUserEmailCommand) (*ChangeEmailResult, error) {

}

func (s *userApplicationService) DeactivateUser(ctx context.Context, cmd ChangeUserStatusCommand) error {

}

func (s *userApplicationService) ActivateUser(ctx context.Context, cmd ChangeUserStatusCommand) error {

}

func (s *userApplicationService) GetActiveUsers(ctx context.Context) ([]*UserDTO, error) {

}
