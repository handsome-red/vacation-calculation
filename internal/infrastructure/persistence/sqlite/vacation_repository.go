package sqlite

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/jmoiron/sqlx"
)

type vacationRepository struct {
	db *sqlx.DB
}

func NewVacationRepository(db *sqlx.DB) *vacationRepository{
	return &vacationRepository{
		db: db,
	}
}

func(r*vacationRepository)Save(ctx context.Context, vacation *vacation.Vacation) error {
	return nil
}

func(r*vacationRepository)FindByUserID(ctx context.Context, userID user.UserID) ( *vacation.Vacation, error) {
	return nil, nil
}


// func(r *vacationRepository) NewVacationRepository()