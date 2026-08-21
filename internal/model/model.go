// internal/model/vacation.go
package model

import "time"

type Vacation struct {
	ID        int       `json:"id"`
	Employee  string    `json:"employee"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Days      int       `json:"days"`
	Status    string    `json:"status"` // planned, approved, completed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VacationRequest struct {
	Employee  string `json:"employee"`
	StartDate string `json:"start_date"`
	Days      int    `json:"days"`
}

type VacationResponse struct {
	ID        int    `json:"id"`
	Employee  string `json:"employee"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Days      int    `json:"days"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}