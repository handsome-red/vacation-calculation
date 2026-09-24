package dto

import "github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"

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

	WorkdayDuration int  `json:"workday_duration" validate:"required,min=1,max=1440"`
	IsInvalid       bool `json:"is_invalid"`
}

func (r RegisterUserRequest) ToCommand() register_user.Command {
	return register_user.Command{
		Email:           r.Email,
		Password:        r.Password,
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		MiddleName:      r.MiddleName,
		Status:          r.Status,
		BirthDate:       r.BirthDate,
		Position:        r.Position,
		HiredAt:         r.HiredAt,
		DistrictCode:    r.DistrictCode,
		DepartmentCode:  r.DepartmentCode,
		WorkdayDuration: r.WorkdayDuration,
		IsInvalid:       r.IsInvalid,
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

	DistrictCode    string `json:"district_code"`
	DistrictTitle   string `json:"district_title"`
	DepartmentCode  string `json:"department_code"`
	DepartmentTitle string `json:"department_title"`

	WorkdayDuration int    `json:"workday_duration"`
	Email           string `json:"email"`
	IsInvalid       bool   `json:"is_invalid"`
	TotalExperience string `json:"total_experience,omitempty"`
}
