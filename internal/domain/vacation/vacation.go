package vacation

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusRequested Status = "REQUESTED"
	StatusApproved  Status = "APPROVED"
	StatusRejected  Status = "REJECTED"
	StatusCancelled Status = "CANCELLED"
)

var (
	ErrInvalidEmployeeID = errors.New("invalid employee id")
	ErrInvalidDateRange  = errors.New("start date must be before end date")
	ErrZeroDays          = errors.New("vacation must be at least 1 day")
	ErrAlreadyProcessed  = errors.New("vacation already processed")
)

type Vacation struct {
	id         string
	employeeID string
	startDate  time.Time
	endDate    time.Time
	days       int
	status     Status
	createdAt  time.Time
	updatedAt  time.Time
	// TODO
}

// Конструктор - создает валидный отпуск
func NewVacation(employeeID string, startDate time.Time, endDate time.Time) (*Vacation, error) {
	if employeeID == "" {
		return nil, ErrInvalidEmployeeID
	}

	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}

	days := calculateDays(startDate, endDate)
	if days <= 0 {
		return nil, ErrZeroDays
	}

	return &Vacation{
		id:         uuid.New().String(),
		employeeID: employeeID,
		startDate:  startDate,
		endDate:    endDate,
		days:       days,
		status:     StatusRequested,
		createdAt:  time.Now(),
		updatedAt:  time.Now(),
	}, nil
}

func calculateDays(startDate, endDate time.Time) int {
	return int(endDate.Sub(startDate).Hours() / 24)
}

func (v *Vacation) Approve() error {

}
