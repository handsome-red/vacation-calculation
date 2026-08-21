package repository

import "vacation-calculation/internal/model"

type VacationRepository interface {
	FindAll() ([]model.Vacation, error)
	FindByID(id int) (*model.Vacation, error)
	Create(vacation *model.Vacation) error
	Update(vacation *model.Vacation) error
	Delete(id int) error
	FindByEmployee(employee string) ([]model.Vacation, error)
}
