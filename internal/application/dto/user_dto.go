package dto

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type UserDTO struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`
}

func NewUserDTOFromDomain(u *user.User) *UserDTO {
	return &UserDTO{
		ID:         u.ID().String(),
		Email:      u.Email().String(),
		FullName:   u.FullName(),
		Department: u.Department().String(),
		CreatedAt:  u.CreatedAt(),
		UpdatedAt:  u.UpdatedAt(),
	}
}

func NewUserDTOsFromDomain(users []*user.User) []*UserDTO {
	dtos := make([]*UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, NewUserDTOFromDomain(u))
	}
	return dtos
}
