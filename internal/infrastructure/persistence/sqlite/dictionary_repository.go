// infrastructure/persistence/sqlite/district_repository.go
package sqlite

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type districtRepository struct {
	db *sqlx.DB
}

type departmentRepository struct {
	db *sqlx.DB
}

func NewDistrictRepository(db *sqlx.DB) user.DistrictRepository {
	return &districtRepository{db: db}
}

func NewDepartmentRepository(db *sqlx.DB) user.DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *districtRepository) List(ctx context.Context) ([]user.District, error) {
	var rows []struct {
		Code  string `db:"code"`
		Title string `db:"title"`
	}
	const q = `SELECT code, title FROM districts WHERE is_active = 1 ORDER BY title`
	if err := r.db.SelectContext(ctx, &rows, q); err != nil {
		return nil, fmt.Errorf("list districts: %w", err)
	}

	result := make([]user.District, 0, len(rows))
	for _, row := range rows {
		result = append(result, user.NewDistrict(row.Code, row.Title))
	}
	return result, nil
}

func (r *departmentRepository) List(ctx context.Context) ([]user.Department, error) {
	var rows []struct {
		Code  string `db:"code"`
		Title string `db:"title"`
	}

	const q = `SELECT code, title FROM departments WHERE is_active = 1 ORDER BY title`
	if err := r.db.SelectContext(ctx, &rows, q); err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	result := make([]user.Department, 0, len(rows))
	for _, row := range rows {
		result = append(result, user.Department{
			Code:  row.Code,
			Title: row.Title,
		})
	}
	return result, nil
}

func (r *departmentRepository) ListWithDistrict(ctx context.Context) ([]user.DepartmentWithDistrict, error) {
	var rows []struct {
		Code         string `db:"code"`
		Title        string `db:"title"`
		DistrictCode string `db:"district_code"`
	}

	const q = `SELECT code, title, district_code FROM departments WHERE is_active = 1 ORDER BY title`
	if err := r.db.SelectContext(ctx, &rows, q); err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	result := make([]user.DepartmentWithDistrict, 0, len(rows))
	for _, row := range rows {
		result = append(result, user.DepartmentWithDistrict{
			Department: user.Department{
				Code:  row.Code,
				Title: row.Title,
			},
			DistrictCode: row.DistrictCode,
		})
	}
	return result, nil
}
