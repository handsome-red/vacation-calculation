package vacation

import (
	"strings"
)

type VacationStatus string

const (
	StatusUnknown VacationStatus = "UNKNOWN"

	StatusDraft     VacationStatus = "DRAFT"
	StatusApproved  VacationStatus = "APPROVED"
	StatusCancelled VacationStatus = "CANCELLED"
	// StatusTransferred — отпуск перенесён на другие даты; новые даты — отдельная запись.
	StatusTransferred VacationStatus = "TRANSFERRED"
)

// TODO
// Отзыв - возврат из отпуска по производной необходимости (?)

var vacationStatusMap = map[VacationStatus]string{
	StatusDraft:       "Черновик",
	StatusApproved:    "Принято",
	StatusCancelled:   "Отменён",
	StatusTransferred: "Перенесён",
}

// IsActive — отпуск действует: занимает даты и дни квоты.
func (v VacationStatus) IsActive() bool {
	return v == StatusDraft || v == StatusApproved
}

func NewVacationStatus(s string) (VacationStatus, error) {
	v := VacationStatus(strings.ToUpper(strings.TrimSpace(s)))
	if !v.isValid() {
		return StatusUnknown, ErrStatusInvalid
	}
	return v, nil
}

func ParseStatus(s string) (VacationStatus, error) {
	v := VacationStatus(strings.ToUpper(strings.TrimSpace(s)))
	if !v.isValid() {
		return StatusUnknown, ErrStatusNotFound
	}
	return v, nil
}

func (v VacationStatus) isValid() bool {
	_, ok := vacationStatusMap[v]
	return ok
}

func (v VacationStatus) String() string {
	return vacationStatusMap[v]
}

func (v VacationStatus) ISO() string {
	return string(v)
}
