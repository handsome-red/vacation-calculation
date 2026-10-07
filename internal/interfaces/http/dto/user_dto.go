package dto

import (
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type UserDTO struct {
	ID              string
	Status          string
	IsActive        bool
	LastName        string
	FirstName       string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	Department      string
	District        string
	WorkdayDuration int
	Email           string
	IsInvalid       bool
	Experience      string
	CreatedAt       string
	UpdatedAt       string

	Initials string
}

func ToUser(u *user.User) UserDTO {
	return UserDTO{
		ID:              u.ID().String(),
		Status:          u.Status().String(),
		IsActive:        u.IsActive(),
		LastName:        u.LastName(),
		FirstName:       u.FirstName(),
		MiddleName:      u.MiddleName(),
		BirthDate:       u.BirthDate().HumanRead(),
		Position:        u.Position().Title(),
		HiredAt:         u.HiredAt().String(),
		Department:      u.Department().Title,
		District:        u.District().Title(),
		WorkdayDuration: u.WorkdayDuration().Int(),
		Email:           u.Email().String(),
		IsInvalid:       u.IsInvalid(),
		Experience:      u.Experience().String(),
		CreatedAt:       u.CreatedAt().Format(dateFormat),
		UpdatedAt:       u.UpdatedAt().Format(dateFormat),
		Initials:        u.Initials(),
	}
}

type RegisterUserRequest struct {
	Email      string `json:"email"           validate:"required,email"`
	Password   string `json:"password"        validate:"required,min=8,max=72"`
	FirstName  string `json:"first_name"      validate:"required"`
	LastName   string `json:"last_name"       validate:"required"`
	MiddleName string `json:"middle_name"`

	Status    string `json:"status"          validate:"required,oneof=active blocked fired"`
	BirthDate string `json:"birth_date"      validate:"required,datetime=2006-01-02"`
	Position  string `json:"position"        validate:"required"`
	HiredAt   string `json:"hired_at"        validate:"required,datetime=2006-01-02"`

	DistrictCode   string `json:"district_code"   validate:"required"`
	DepartmentCode string `json:"department_code" validate:"required"`
}

func (r RegisterUserRequest) ToCommand() register_user.Command {
	return register_user.Command{
		Email:          r.Email,
		Password:       r.Password,
		FirstName:      r.FirstName,
		LastName:       r.LastName,
		MiddleName:     r.MiddleName,
		BirthDate:      r.BirthDate,
		Position:       r.Position,
		HiredAt:        r.HiredAt,
		DistrictCode:   r.DistrictCode,
		DepartmentCode: r.DepartmentCode,
	}
}

type UserResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	BirthDate  string `json:"birth_date"`
	Position   string `json:"position"`
	HiredAt    string `json:"hired_at"`

	// DistrictCode    string `json:"district_code"`
	DistrictTitle string `json:"district_title"`
	// DepartmentCode  string `json:"department_code"`
	DepartmentTitle string `json:"department_title"`

	Email           string `json:"email"`
	TotalExperience string `json:"total_experience,omitempty"`
}
