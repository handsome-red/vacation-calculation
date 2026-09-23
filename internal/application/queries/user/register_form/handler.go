package register_form

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Handle(ctx context.Context, _ Query) (*Result, error) {

	positions := user.AllPositions()
	districts := user.AllDistricts()
	departments := user.AllDepartments()

	pos := make([]PositionOption, 0, len(positions))
	for _, p := range positions {
		pos = append(pos, PositionOption{
			Value: p.String(),
			Title: p.Title(),
		})
	}

	dis := make([]DistrictsOption, 0, len(districts))
	for _, d := range districts {
		dis = append(dis, DistrictsOption{
			Value: d.String(),
			Title: d.Title(),
		})
	}

	dep := make([]DepartmentOption, 0, len(departments))
	for _, d := range departments {
		dep = append(dep, DepartmentOption{
			Value: d.String(),
			Title: d.Title(),
		})
	}

	return &Result{
		Positions:   pos,
		Districts:   dis,
		Departments: dep,
	}, nil
}
