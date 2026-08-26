package types

type EmployeeStatus string

const (
	StatusActive   EmployeeStatus = "active"
	StatusVacation EmployeeStatus = "vacation"
	StatusFired    EmployeeStatus = "fired"
)

type WorkDayDuration int

const (
	WorkDayFull WorkDayDuration = 8
)
