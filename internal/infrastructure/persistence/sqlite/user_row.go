package sqlite

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

const sqliteTimeLayout = "2006-01-02 15:04:05"

type userRow struct {
	ID              string `db:"id"`
	Status          string `db:"status"`
	LastName        string `db:"last_name"`
	FirstName       string `db:"first_name"`
	MiddleName      string `db:"middle_name"`
	BirthDate       string `db:"birth_date"`
	Position        string `db:"position"`
	HiredAt         string `db:"hired_at"`
	Department      string `db:"department"`
	Email           string `db:"email"`
	IsInvalid       int    `db:"is_invalid"`
	Password        string `db:"password"`
	CreatedAt       string `db:"created_at"`
	UpdatedAt       string `db:"updated_at"`
	District        string `db:"district"`
	WorkdayDuration int    `db:"workday_duration"`
}

func (r *userRow) toDomain() (*user.User, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	userID, err := user.NewUserID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	status, err := user.NewStatus(r.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	email, err := user.NewEmail(r.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	birthDate, err := user.NewBirthDate(r.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("invalid birth_date: %w", err)
	}

	position, err := user.NewPosition(r.Position)
	if err != nil {
		return nil, fmt.Errorf("invalid position: %w", err)
	}

	hiredAt, err := user.NewHiredDate(r.HiredAt)
	if err != nil {
		return nil, fmt.Errorf("invalid hired_at: %w", err)
	}

	department, err := user.NewDepartment(r.Department)
	if err != nil {
		return nil, fmt.Errorf("invalid department: %w", err)
	}

	district, err := user.NewDistrict(r.District)
	if err != nil {
		return nil, fmt.Errorf("invalid district: %w", err)
	}

	workdayDuration, err := user.NewWorkdayDuration(r.WorkdayDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid workday_duration: %w", err)
	}

	createdAt, err := parseSQLiteTime(r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}

	updatedAt, err := parseSQLiteTime(r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at: %w", err)
	}

	return user.ReconstructUser(user.UserParams{
		ID:              userID,
		Status:          status,
		Email:           email,
		Password:        user.NewPasswordFromHash(r.Password),
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		MiddleName:      r.MiddleName,
		BirthDate:       birthDate,
		Position:        position,
		HiredAt:         hiredAt,
		Department:      department,
		District:        district,
		WorkdayDuration: workdayDuration,
		IsInvalid:       r.IsInvalid != 0,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}), nil
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
