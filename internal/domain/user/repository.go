package user

import "context"

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, userID UserID) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Delete(ctx context.Context, userID UserID) error
}

type DistrictRepository interface {
	List(ctx context.Context) ([]District, error)
}

type DepartmentRepository interface {
	List(ctx context.Context) ([]Department, error)
	ListWithDistrict(ctx context.Context) ([]DepartmentWithDistrict, error)
}
