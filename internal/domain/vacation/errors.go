package vacation

import "errors"

var (
	ErrIDEmpty     = errors.New("id is empty")
	ErrInvalidDate = errors.New("date invalid")

	ErrVacationIDEmpty    = errors.New("vacation id is empty")
	ErrUserIDEmpty        = errors.New("user id is empty")
	ErrStartDateEmpty     = errors.New("start date is empty")
	ErrEndDateEmpty       = errors.New("end date is empty")
	ErrEndDateBeforeStart = errors.New("end date is before start date")
	ErrStatusInvalid      = errors.New("status is invalid")

	ErrHolidayDateEmpty = errors.New("holiday date is empty")
	ErrHolidayNameEmpty = errors.New("holiday name is empty")

	ErrStatusNotFound = errors.New("status not found")

	ErrVacationNotRejectable = errors.New("vacation is not rejectable")
	ErrVacationNotEditable   = errors.New("vacation is not editable")
	ErrVacationNotApprovable = errors.New("vacation is not approvable")

	ErrInvalidEmployeeID = errors.New("invalid employee id")
	ErrInvalidDateRange  = errors.New("start date must be before end date")
	ErrZeroDays          = errors.New("vacation must be at least 1 day")
	ErrAlreadyProcessed  = errors.New("vacation already processed")

	ErrCalcInvalidTime = errors.New("hired after now")
	ErrCalcEmptyStats  = errors.New("empty status")
)
