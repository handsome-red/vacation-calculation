package sqlite

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

const sqliteTimeLayout = "2006-01-02 15:04:05"

type userRow struct {
	ID              string
	Status          string
	LastName        string
	FirstName       string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	Department      string
	Email           string
	IsInvalid       int
	Password        string
	CreatedAt       string
	UpdatedAt       string
	District        string
	WorkdayDuration int
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

	createdAt, err := parseSQLiteTime(r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}
	updatedAt, err := parseSQLiteTime(r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at: %w", err)
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
		createdAt,
		updatedAt,
	), nil
}

func parseSQLiteTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(sqliteTimeLayout, s)
	if err != nil {
		// На случай, если формат другой (RFC3339 или с миллисекундами)
		if t2, err2 := time.Parse(time.RFC3339, s); err2 == nil {
			return t2.UTC(), nil
		}
		return time.Time{}, fmt.Errorf("parse sqlite time %q: %w", s, err)
	}
	return t.UTC(), nil
}
