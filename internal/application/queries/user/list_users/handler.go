package list_users

import (
	"context"
	"fmt"
	"sort"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
	userRepo       ports.UserRepository
	districtRepo   ports.DistrictRepository
	departmentRepo ports.DepartmentRepository
	logger         ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	districtRepo ports.DistrictRepository,
	departmentRepo ports.DepartmentRepository,
	logger ports.Logger,
) *Handler {
	return &Handler{
		userRepo:       userRepo,
		districtRepo:   districtRepo,
		departmentRepo: departmentRepo,
		logger:         logger,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	// Дефолтные значения
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Size < 1 || query.Size > 100 {
		query.Size = 20
	}

	f := query.Filters
	filter := ports.UserFilter{
		LastName:       f.LastName,
		FirstName:      f.FirstName,
		MiddleName:     f.MiddleName,
		Email:          f.Email,
		Status:         f.Status,
		Position:       f.Position,
		DistrictCode:   f.DistrictCode,
		DepartmentCode: f.DepartmentCode,
	}
	if f.Category != "" {
		filter.Positions = user.PositionsOfCategory(f.Category)
	}

	users, total, err := h.userRepo.FindUsers(ctx, ports.Query{Page: query.Page, Size: query.Size, Filters: filter})
	if err != nil {
		h.logger.Error(ctx, "failed to get users", "error", err)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	result := &Result{
		Users: make([]UserListItem, 0, len(users)),
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
		Pages: (total + query.Size - 1) / query.Size,
	}

	for _, u := range users {
		result.Users = append(result.Users, UserListItem{
			ID:              u.ID().String(),
			Status:          u.Status().Label(),
			Email:           u.Email().String(),
			FirstName:       u.FirstName(),
			LastName:        u.LastName(),
			MiddleName:      u.MiddleName(),
			BirthDate:       u.BirthDate().String(),
			Position:        u.Position().Title(),
			Category:        user.CategoryTitle(u.Position().Category()),
			HiredAt:         u.HiredAt().String(),
			DepartmentTitle: u.Department().Title,
			DistrictTitle:   u.District().Title(),
			WorkdayDuration: u.WorkdayDuration().Int(),
			IsInvalid:       u.IsInvalid(),
			TotalExperience: u.ExperienceLabel(),
		})
	}

	// Справочники для фильтров.
	districts, err := h.districtRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("districts: %w", err)
	}
	for _, d := range districts {
		result.Districts = append(result.Districts, Option{Code: d.Code(), Title: d.Title()})
	}
	departments, err := h.departmentRepo.ListWithDistrict(ctx)
	if err != nil {
		return nil, fmt.Errorf("departments: %w", err)
	}
	for _, d := range departments {
		result.Departments = append(result.Departments, DepartmentOption{Code: d.Department.Code, Title: d.Department.Title, DistrictCode: d.DistrictCode})
	}
	for _, p := range user.AllPositions() {
		result.Positions = append(result.Positions, Option{Code: string(p), Title: p.Title()})
	}
	sort.Slice(result.Positions, func(i, j int) bool { return result.Positions[i].Title < result.Positions[j].Title })
	for _, c := range []string{user.CategoryCivilServant, user.CategoryOtherStaff} {
		result.Categories = append(result.Categories, Option{Code: c, Title: user.CategoryTitle(c)})
	}

	return result, nil
}
