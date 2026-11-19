package update_user

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
	userRepo       ports.UserRepository
	districtRepo   ports.DistrictRepository
	departmentRepo ports.DepartmentRepository
	vacationRepo   ports.VacationRepository
	openingRepo    ports.OpeningUsageRepository
	workYearRepo   ports.WorkYearRepository
	hasher         ports.PasswordHasher
}

func NewHandler(
	userRepo ports.UserRepository,
	districtRepo ports.DistrictRepository,
	departmentRepo ports.DepartmentRepository,
	vacationRepo ports.VacationRepository,
	openingRepo ports.OpeningUsageRepository,
	workYearRepo ports.WorkYearRepository,
	hasher ports.PasswordHasher,
) *Handler {
	return &Handler{
		userRepo:       userRepo,
		districtRepo:   districtRepo,
		departmentRepo: departmentRepo,
		vacationRepo:   vacationRepo,
		openingRepo:    openingRepo,
		workYearRepo:   workYearRepo,
		hasher:         hasher,
	}
}

func invalid(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}

// Handle заменяет данные сотрудника. Дату приёма нельзя менять, пока есть действующие
// отпуска: они привязаны к началу рабочего года, и новые границы лет отвязали бы их
// от квот — уже списанные дни снова стали бы доступны.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	id, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("%w: parse user id: %v", user.ErrUserNotFound, err)
	}
	u, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !cmd.Caller.Manages(u) {
		return ErrForbidden
	}

	email, err := user.NewEmail(cmd.Email)
	if err != nil {
		return invalid("некорректный e-mail")
	}
	other, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return fmt.Errorf("check email: %w", err)
	}
	if err == nil && other.ID() != u.ID() {
		return invalid("этот e-mail уже используется другим сотрудником")
	}

	birthDate, err := user.NewBirthDate(cmd.BirthDate)
	if err != nil {
		return invalid("некорректная дата рождения")
	}
	position, err := user.NewPosition(cmd.Position)
	if err != nil {
		return invalid("выберите должность")
	}
	hiredAt, err := user.NewHiredDate(cmd.HiredAt)
	if err != nil {
		return invalid("некорректная дата приёма на работу")
	}
	workday, err := user.NewWorkdayDuration(cmd.WorkdayDuration)
	if err != nil {
		return invalid("продолжительность рабочего дня — от 1 до 8 часов")
	}
	district, err := h.districtRepo.FindByCode(ctx, cmd.DistrictCode)
	if err != nil {
		return invalid("выберите территориальный отдел")
	}
	if !cmd.Caller.ManagesDistrict(district.Code()) {
		return invalid("можно выбрать только территориальный отдел, которым вы руководите")
	}
	department, err := h.findDepartment(ctx, cmd.DepartmentCode, cmd.DistrictCode)
	if err != nil {
		return err
	}

	hiredChanged := !hiredAt.Time().Equal(u.HiredAt().Time())
	if hiredChanged {
		vacations, err := h.vacationRepo.FindByUserID(ctx, u.ID())
		if err != nil {
			return fmt.Errorf("find vacations: %w", err)
		}
		active := 0
		for _, v := range vacations {
			if v.Status().IsActive() {
				active++
			}
		}
		opening, err := h.openingRepo.ListByUser(ctx, u.ID())
		if err != nil {
			return fmt.Errorf("list opening usage: %w", err)
		}
		if len(opening) > 0 {
			return invalid("нельзя изменить дату приёма: к рабочим годам привязаны дни, использованные до начала учёта")
		}
		years, err := h.workYearRepo.ListByUser(ctx, u.ID())
		if err != nil {
			return fmt.Errorf("list work years: %w", err)
		}
		for _, y := range years {
			if y.Manual {
				return invalid("нельзя изменить дату приёма: границы рабочих лет заданы вручную. Сначала верните их расчёт")
			}
		}
		if active > 0 {
			return invalid(fmt.Sprintf(
				"нельзя изменить дату приёма: у сотрудника %d действующих записей отпуска, привязанных к рабочим годам. Сначала отмените их", active))
		}
	}

	var newPassword *user.Password
	if cmd.Password != "" {
		if n := utf8.RuneCountInString(cmd.Password); n < 8 || len(cmd.Password) > 72 {
			return invalid("пароль — от 8 символов (не больше 72 байт)")
		}
		hash, err := h.hasher.HashPassword(ctx, cmd.Password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		p := user.NewPasswordFromHash(hash)
		newPassword = &p
	}

	err = u.UpdateProfile(user.Profile{
		LastName:        cmd.LastName,
		FirstName:       cmd.FirstName,
		MiddleName:      cmd.MiddleName,
		BirthDate:       birthDate,
		Position:        position,
		HiredAt:         hiredAt,
		ExperienceDays:  cmd.Experience,
		Department:      department,
		District:        district,
		WorkdayDuration: workday,
		Email:           email,
		IsInvalid:       cmd.IsInvalid,
	})
	switch {
	case errors.Is(err, user.ErrNameRequired):
		return invalid("укажите фамилию, имя и отчество")
	case errors.Is(err, user.ErrExperienceInvalid):
		return invalid("стаж — целое неотрицательное число дней")
	case err != nil:
		return err
	}

	if newPassword != nil {
		if err := u.ChangePassword(*newPassword); err != nil {
			return fmt.Errorf("change password: %w", err)
		}
	}

	// Сохранённые годы считаны от прежней даты приёма: они сохранятся заново от новой.
	// Удаляются до сохранения сотрудника — при ошибке сохранения годы просто посчитаются
	// от прежней даты.
	if hiredChanged {
		if err := h.workYearRepo.DeleteByUser(ctx, u.ID()); err != nil {
			return fmt.Errorf("delete work years: %w", err)
		}
	}

	if err := h.userRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

// findDepartment — район/отдел из справочника: республиканский либо этого теротдела.
func (h *Handler) findDepartment(ctx context.Context, code, districtCode string) (user.Department, error) {
	departments, err := h.departmentRepo.ListWithDistrict(ctx)
	if err != nil {
		return user.Department{}, fmt.Errorf("list departments: %w", err)
	}
	for _, d := range departments {
		if d.Department.Code != code {
			continue
		}
		if d.DistrictCode != "" && d.DistrictCode != districtCode {
			return user.Department{}, invalid("район/отдел не относится к выбранному территориальному отделу")
		}
		return d.Department, nil
	}
	return user.Department{}, invalid("выберите район/отдел")
}
