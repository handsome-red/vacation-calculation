package sqlite

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type userRow struct {
	ID              string `db:"id"`
	Status          string `db:"status"`
	LastName        string `db:"last_name"`
	FirstName       string `db:"first_name"`
	MiddleName      string `db:"middle_name"`
	BirthDate       string `db:"birth_date"`
	Position        string `db:"position"`
	HiredAt         string `db:"hired_at"`
	DepartmentCode  string `db:"department_code"`
	DepartmentTitle string `db:"department_title"`
	DistrictCode    string `db:"district_code"`
	DistrictTitle   string `db:"district_title"`
	WorkdayDuration int    `db:"workday_duration"`
	Email           string `db:"email"`
	IsInvalid       int    `db:"is_invalid"`
	Password        string `db:"password"`
	CreatedAt       string `db:"created_at"`
	UpdatedAt       string `db:"updated_at"`
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

	department := user.Department{
		Code:  r.DepartmentCode,
		Title: r.DepartmentTitle,
	}
	district := user.NewDistrict(r.DistrictCode)

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
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("parse sqlite time %q", s)
}
