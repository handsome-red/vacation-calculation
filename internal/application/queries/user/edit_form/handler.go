package edit_form

import (
	"context"
	"fmt"
	"strconv"

	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/register_form"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Query struct {
	UserID string
	// Caller — кто изменяет: руководитель выбирает только свой теротдел.
	Caller auth.Caller
}

// Form — значения полей формы как строки: при ошибке показываются введённые значения.
type Form struct {
	LastName        string
	FirstName       string
	MiddleName      string
	Email           string
	BirthDate       string
	Position        string
	Experience      string
	HiredAt         string
	District        string
	Department      string
	WorkdayDuration string
	IsInvalid       bool
}

type Result struct {
	UserID      string
	Positions   []register_form.PositionOption
	Districts   []register_form.DistrictOption
	Departments []register_form.DepartmentOption
	Today       string
	Form        Form
	Error       string
}

type Handler struct {
	userRepo ports.UserRepository
	options  *register_form.Handler
}

func NewHandler(userRepo ports.UserRepository, options *register_form.Handler) *Handler {
	return &Handler{userRepo: userRepo, options: options}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	id, err := user.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: parse user id: %v", user.ErrUserNotFound, err)
	}
	u, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	options, err := h.options.Handle(ctx, register_form.Query{Districts: q.Caller.Scope()})
	if err != nil {
		return nil, err
	}

	return &Result{
		UserID:      u.ID().String(),
		Positions:   options.Positions,
		Districts:   options.Districts,
		Departments: options.Departments,
		Today:       options.Today,
		Form: Form{
			LastName:        u.LastName(),
			FirstName:       u.FirstName(),
			MiddleName:      u.MiddleName(),
			Email:           u.Email().Value(),
			BirthDate:       u.BirthDate().ISO(),
			Position:        u.Position().String(),
			Experience:      strconv.Itoa(u.ExperienceDays()),
			HiredAt:         u.HiredAt().Time().Format("2006-01-02"),
			District:        u.District().Code(),
			Department:      u.Department().Code,
			WorkdayDuration: strconv.Itoa(u.WorkdayDuration().Int()),
			IsInvalid:       u.IsInvalid(),
		},
	}, nil
}
