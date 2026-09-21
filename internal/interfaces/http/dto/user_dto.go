package dto

import "github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"

type RegisterUserRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8,max=72"`
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	MiddleName string `json:"middle_name"`
	Department string `json:"department" validate:"required"`
}

func (r RegisterUserRequest) ToCommand() register_user.Command {
	return register_user.Command{
		Email:      r.Email,
		Password:   r.Password,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		MiddleName: r.MiddleName,
		Department: r.Department,
	}
}

type UserResponse struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	LastName        string `json:"last_name"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name,omitempty"`
	BirthDate       string `json:"birth_date"`
	Position        string `json:"position"`
	HiredAt         string `json:"hired_at"`
	Department      string `json:"department"`
	District        string `json:"district"`
	WorkdayDuration int    `json:"workday_duration"`
	Email           string `json:"email"`
	IsInvalid       bool   `json:"is_invalid"`
	TotalExperience string `json:"total_experience"`
}
