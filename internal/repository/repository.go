package repository

import "github.com/handsome-red/vacation-calculation/internal/model"

type VacationRepository interface {
	Ping() error
	FindAll() ([]model.Vacation, error)
	FindByID(id int) (*model.Vacation, error)
	Create(vacation *model.Vacation) error
	Update(vacation *model.Vacation) error
	Delete(id int) error
	FindByUserID(userID string) ([]model.Vacation, error)
}
