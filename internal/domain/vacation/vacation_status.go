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
)

// TODO
// Перенос
// Отзыв - возврат из отпуска по производной необходимости (?)

// Принято
// Отменен
// Перенос (?)
var vacationStatusMap = map[VacationStatus]string{
	StatusDraft:     "Черновик",
	StatusApproved:  "Принято",
	StatusCancelled: "Отменён",
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
