package vacation

import (
	"strings"
)

type VacationStatus string

const (
	StatusUnknown VacationStatus = "UNKNOWN"

	StatusDraft VacationStatus = "DRAFT"

	StatusRequested VacationStatus = "REQUESTED"
	StatusApproved  VacationStatus = "APPROVED"
	StatusRejected  VacationStatus = "REJECTED"
	StatusCancelled VacationStatus = "CANCELLED"
)

// TODO
var vacationStatusMap = map[VacationStatus]string{
	StatusDraft:     "Черновик",
	StatusApproved:  "Принято",
	StatusRejected:  "Отказано",
	StatusCancelled: "Подтвержден",
}

func NewVacationStatus(s string) (VacationStatus, error) {
	v := VacationStatus(strings.ToLower(strings.TrimSpace(s)))
	if !v.isValid() {
		return StatusUnknown, ErrStatusInvalid
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
