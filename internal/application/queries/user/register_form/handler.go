package register_form

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
	districtRepo   ports.DistrictRepository
	departmentRepo ports.DepartmentRepository
}

func NewHandler(
	districtRepo ports.DistrictRepository,
	departmentRepo ports.DepartmentRepository,
) *Handler {
	return &Handler{
		districtRepo:   districtRepo,
		departmentRepo: departmentRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, _ Query) (*Result, error) {

	positions := user.AllPositions()
	districts, err := h.districtRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("districts repo: %w", err)
	}

	departments, err := h.departmentRepo.ListWithDistrict(ctx)
	if err != nil {
		return nil, fmt.Errorf("departments repo: %w", err)
	}

	pos := make([]PositionOption, 0, len(positions))
	for _, p := range positions {
		pos = append(pos, PositionOption{
			Code:  p.String(),
			Title: p.Title(),
		})
	}

	dis := make([]DistrictOption, 0, len(districts))
	for _, d := range districts {
		dis = append(dis, DistrictOption{
			Code:  d.Code(),
			Title: d.Title(),
		})
	}

	dep := make([]DepartmentOption, 0, len(departments))
	for _, d := range departments {
		dep = append(dep, DepartmentOption{
			Code:         d.Department.Code,
			Title:        d.Department.Title,
			DistrictCode: d.DistrictCode,
		})
	}

	return &Result{
		Positions:   pos,
		Districts:   dis,
		Departments: dep,
	}, nil
}
