package sqlite

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type userRow struct {
	ID         string
	Email      string
	Password   string
	FirstName  string
	LastName   string
	MiddleName string
	Department string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r *userRow) toDomain() (*user.User, error) {
	parsedID, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID %q: %w", r.ID, err)
	}

	id, err := user.NewUserID(parsedID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID %q: %w", r.ID, err)
	}

	email, err := user.NewEmail(r.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email %q: %w", r.Email, err)
	}

	department, err := user.NewDepartment(r.Department)
	if err != nil {
		return nil, fmt.Errorf("invalid department %q: %w", r.Department, err)
	}

	password := user.NewPasswordFromHash(r.Password)

	return user.ReconstructUser(
		id,
		email,
		password,
		r.FirstName,
		r.LastName,
		r.MiddleName,
		department,
		r.CreatedAt,
		r.UpdatedAt,
	), nil
}
