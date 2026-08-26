package service

import (
	"errors"
	"time"
	"vacation-calculation/internal/model"
	"vacation-calculation/internal/repository"
)

type VacationService interface {
	Ping() error
	GetAllVacations() ([]model.Vacation, error)
	GetVacationByID(id int) (*model.Vacation, error)
	CreateVacation(req model.VacationRequest) (*model.Vacation, error)
	UpdateVacation(id int, req model.VacationRequest) (*model.Vacation, error)
	DeleteVacation(id int) error
	GetVacationsByEmployee(employee string) ([]model.Vacation, error)
	CalculateVacation(days int, startDate string) (time.Time, error)
}

type vacationService struct {
	repo repository.VacationRepository
}

func NewVacationService(repo repository.VacationRepository) VacationService {
	return &vacationService{
		repo: repo,
	}
}

func (s *vacationService) Ping() error {
	return s.repo.Ping()
}

func (s *vacationService) GetAllVacations() ([]model.Vacation, error) {
	return s.repo.FindAll()
}

func (s *vacationService) GetVacationByID(id int) (*model.Vacation, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.repo.FindByID(id)
}

func (s *vacationService) CreateVacation(req model.VacationRequest) (*model.Vacation, error) {
	// Валидация
	if req.Employee == "" {
		return nil, errors.New("employee name is required")
	}
	if req.Days <= 0 || req.Days > 365 {
		return nil, errors.New("days must be between 1 and 365")
	}

	// Парсим дату
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	// Рассчитываем дату окончания
	endDate := startDate.AddDate(0, 0, req.Days)

	// Создаем отпуск
	vacation := &model.Vacation{
		Employee:  req.Employee,
		StartDate: startDate,
		EndDate:   endDate,
		Days:      req.Days,
		Status:    "planned",
	}

	if err := s.repo.Create(vacation); err != nil {
		return nil, err
	}

	return vacation, nil
}

func (s *vacationService) UpdateVacation(id int, req model.VacationRequest) (*model.Vacation, error) {
	// Проверяем существование
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	if req.Employee != "" {
		existing.Employee = req.Employee
	}
	if req.Days > 0 && req.Days <= 365 {
		existing.Days = req.Days

		// Пересчитываем дату окончания
		if req.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", req.StartDate)
			if err == nil {
				existing.StartDate = startDate
				existing.EndDate = startDate.AddDate(0, 0, req.Days)
			}
		}
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *vacationService) DeleteVacation(id int) error {
	return s.repo.Delete(id)
}

func (s *vacationService) GetVacationsByEmployee(employee string) ([]model.Vacation, error) {
	return s.repo.FindByEmployee(employee)
}

func (s *vacationService) CalculateVacation(days int, startDate string) (time.Time, error) {
	if days <= 0 || days > 365 {
		return time.Time{}, errors.New("days must be between 1 and 365")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, errors.New("invalid date format")
	}

	return start.AddDate(0, 0, days), nil
}
