package ports

import (
	"context"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	// "github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type UserFilter struct {
	Status          string
	FirstName       string
	LastName        string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	Email           string
	DistrictTitle   string
	DepartmentTitle string
	TotalExperience string
}

type Query struct {
	Page    int
	Size    int
	Filters UserFilter
}

type UserCommandRepository interface {
	Save(ctx context.Context, user *user.User) error
	Delete(ctx context.Context, id user.UserID) error
}

type UserQueryRepository interface {
	FindByID(ctx context.Context, id user.UserID) (*user.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindUsers(ctx context.Context, q Query) ([]*user.User, int, error)
}

type UserRepository interface {
	UserCommandRepository
	UserQueryRepository
}

type DistrictRepository interface {
	List(ctx context.Context) ([]user.District, error)
}

type DepartmentRepository interface {
	List(ctx context.Context) ([]user.Department, error)
	ListWithDistrict(ctx context.Context) ([]user.DepartmentWithDistrict, error)
}

type ShiftRepository interface {
	ListByUser(ctx context.Context, userID user.UserID) ([]vacation.Shift, error)
	ListByUserInRange(ctx context.Context, userID user.UserID, from, to time.Time) ([]vacation.Shift, error)
	Save(ctx context.Context, userID user.UserID, shift vacation.Shift) error
}

type VacationRepository interface {
	Save(ctx context.Context, vacation *vacation.Vacation) error
	FindByUserID(ctx context.Context, userID user.UserID) ([]*vacation.Vacation, error)
	FindByUserIDInRange(
		ctx context.Context,
		userID user.UserID,
		from, to vacation.Date,
	) ([]*vacation.Vacation, error)
}

type HolidayRepository interface {
	Save(ctx context.Context, h vacation.Holiday) error
	ListInRange(ctx context.Context, from, to vacation.Date) ([]vacation.Holiday, error)
	ListByYear(ctx context.Context, year int) ([]vacation.Holiday, error)
}

type AuthRepository interface {
	FindByEmail(ctx context.Context, email auth.Email) (*auth.AuthUser, error)
	Save(ctx context.Context, u *auth.AuthUser) error
}

type SessionRepository interface {
	UserBySession(ctx context.Context, sessionID string) (*auth.AuthUser, error)
}
