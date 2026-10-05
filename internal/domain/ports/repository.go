package ports

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
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
