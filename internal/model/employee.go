package model

import (
	// "fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/types"
)

type Employee struct {
	// === Основные поля ===
	ID         int64                `json:"id" db:"id"`
	Status     types.EmployeeStatus `json:"status" db:"status"`
	LastName   string               `json:"last_name" db:"last_name"`
	FirstName  string               `json:"first_name" db:"first_name"`
	MiddleName string               `json:"middle_name" db:"middle_name"`
	Birthday   time.Time            `json:"birthday" db:"birthday"`
	Position   string               `json:"position" db:"position"`
	HireDate   time.Time            `json:"hire_date" db:"hire_date"`

	// === Структурные подразделения ===
	TerritorialDepartment string `json:"territorial_department" db:"territorial_department"` // Территориальный отдел
	Department            string `json:"department" db:"department"`                         // Район/Отдел

	// === Рабочие параметры ===
	WorkDayDuration  types.WorkDayDuration `json:"work_day_duration" db:"work_day_duration"` // Продолжительность рабочего дня
	Email            string                `json:"email" db:"email"`
	IsDisabled       bool                  `json:"is_disabled" db:"is_disabled"`               // Инвалид
	TotalServiceDays int                   `json:"total_service_days" db:"total_service_days"` // Общий стаж выслуги (дни)

	// === Системные поля ===
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewEmployee() *Employee {
	return &Employee{}
}

type EmployeeCreate struct {
}

// func NewEmployeeFromCreate(req *EmployeeCreate) (*Employee, error) {
// 	birthday, err := time.Parse("02.01.2006", req.Birthday)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid birthday format (use DD.MM.YYYY): %w", err)
// 	}

// 	hireDate, err := time.Parse("02.01.2006", req.HireDate)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid hire_date format (use DD.MM.YYYY): %w", err)
// 	}

// 	return NewEmployeeBuilder().
// 		WithStatus(req.Status).
// 		WithLastName(req.LastName).
// 		WithFirstName(req.FirstName).
// 		WithMiddleName(req.MiddleName).
// 		WithBirthday(birthday).
// 		WithPosition(req.Position).
// 		WithHireDate(hireDate).
// 		WithTerritorialDepartment(req.TerritorialDepartment).
// 		WithDepartment(req.Department).
// 		WithWorkDayDuration(req.WorkDayDuration).
// 		WithEmail(req.Email).
// 		WithIsDisabled(req.IsDisabled).
// 		WithTotalServiceDays(req.TotalServiceDays).
// 		Build()
// }
