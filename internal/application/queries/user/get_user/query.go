package get_user

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Query struct {
	UserID string
}

// type Result struct {
// 	ID              string
// 	Status          string
// 	IsActive        bool
// 	LastName        string
// 	FirstName       string
// 	MiddleName      string
// 	BirthDate       string
// 	Position        string
// 	HiredAt         string
// 	Department      string
// 	District        string
// 	WorkdayDuration int
// 	Email           string
// 	IsInvalid       bool
// 	Experience      string
// 	CreatedAt       string
// 	UpdatedAt       string

// 	Initials string
// 	Today    string
// 	WorkYear string
// 	Supposed string
// 	Earned   float32

// 	YearStats []dto.YearStat
// }

type Result struct {
	User      *user.User
	YearStats []vacation.YearStat
	Now       time.Time
}
