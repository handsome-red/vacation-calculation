package model

import (
	"fmt"
	"strings"
	"time"
	"vacation-calculation/internal/types"
)

type EmployeeBuilder struct {
	employee *Employee
	errors   []string
}

func NewEmployeeBuilder() *EmployeeBuilder {
	return &EmployeeBuilder{
		employee: NewEmployee(),
		errors:   []string{},
	}
}

func (b *EmployeeBuilder) WithStatus(status types.EmployeeStatus) *EmployeeBuilder {
	b.employee.Status = status
	return b
}

func (b *EmployeeBuilder) WithLastName(lastName string) *EmployeeBuilder {
	b.employee.LastName = lastName
	return b
}

func (b *EmployeeBuilder) WithFirstName(firstName string) *EmployeeBuilder {
	b.employee.FirstName = firstName
	return b
}

func (b *EmployeeBuilder) WithMiddleName(middleName string) *EmployeeBuilder {
	b.employee.MiddleName = middleName
	return b
}

func (b *EmployeeBuilder) WithBirthday(birthday time.Time) *EmployeeBuilder {
	b.employee.Birthday = birthday
	return b
}

// WithBirthdayString - установка даты рождения из строки (формат: DD.MM.YYYY)
func (b *EmployeeBuilder) WithBirthdayString(birthday string) *EmployeeBuilder {
	parsed, err := time.Parse("02.01.2006", birthday)
	if err != nil {
		b.errors = append(b.errors, fmt.Sprintf("invalid birthday format: %s (use DD.MM.YYYY)", birthday))
	} else {
		b.employee.Birthday = parsed
	}
	return b
}

func (b *EmployeeBuilder) WithPosition(position string) *EmployeeBuilder {
	b.employee.Position = position
	return b
}

func (b *EmployeeBuilder) WithHireDate(hireDate time.Time) *EmployeeBuilder {
	b.employee.HireDate = hireDate
	return b
}

// WithHireDateString - установка даты приема из строки (формат: DD.MM.YYYY)
func (b *EmployeeBuilder) WithHireDateString(hireDate string) *EmployeeBuilder {
	parsed, err := time.Parse("02.01.2006", hireDate)
	if err != nil {
		b.errors = append(b.errors, fmt.Sprintf("invalid hire_date format: %s (use DD.MM.YYYY)", hireDate))
	} else {
		b.employee.HireDate = parsed
	}
	return b
}

func (b *EmployeeBuilder) WithTerritorialDepartment(territorialDepartment string) *EmployeeBuilder {
	b.employee.TerritorialDepartment = territorialDepartment
	return b
}

func (b *EmployeeBuilder) WithDepartment(department string) *EmployeeBuilder {
	b.employee.Department = department
	return b
}

func (b *EmployeeBuilder) WithWorkDayDuration(workDayDuration types.WorkDayDuration) *EmployeeBuilder {
	b.employee.WorkDayDuration = workDayDuration
	return b
}

func (b *EmployeeBuilder) WithEmail(email string) *EmployeeBuilder {
	b.employee.Email = email
	return b
}

func (b *EmployeeBuilder) WithIsDisabled(isDisabled bool) *EmployeeBuilder {
	b.employee.IsDisabled = isDisabled
	return b
}

func (b *EmployeeBuilder) WithTotalServiceDays(totalServiceDays int) *EmployeeBuilder {
	b.employee.TotalServiceDays = totalServiceDays
	return b
}

func (b *EmployeeBuilder) Build() (*Employee, error) {
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("validation errors: %s", strings.Join(b.errors, ";"))
	}

	if err := b.validate(); err != nil {
		return nil, err
	}

	return b.employee, nil
}

func (b *EmployeeBuilder) validate() error {
	var errs []string

	// Проверка обязательных полей
	if b.employee.LastName == "" {
		errs = append(errs, "last_name is required")
	}
	if b.employee.FirstName == "" {
		errs = append(errs, "first_name is required")
	}
	if b.employee.Email == "" {
		errs = append(errs, "email is required")
	}
	if b.employee.Birthday.IsZero() {
		errs = append(errs, "birthday is required")
	}
	if b.employee.HireDate.IsZero() {
		errs = append(errs, "hire_date is required")
	}
	if b.employee.Position == "" {
		errs = append(errs, "position is required")
	}
	if b.employee.Status == "" {
		errs = append(errs, "status is required")
	}
	if string(b.employee.WorkDayDuration) == "" {
		errs = append(errs, "work_day_duration is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}
