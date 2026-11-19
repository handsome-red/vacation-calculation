package update_user

import (
	"errors"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
)

// ErrInvalidInput — данные формы некорректны; текст ошибки показывается администратору.
var ErrInvalidInput = errors.New("неверные данные")

// ErrForbidden — вызывающий не управляет этим сотрудником.
var ErrForbidden = errors.New("доступ запрещён")

type Command struct {
	UserID          string
	LastName        string
	FirstName       string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	Experience      int
	DepartmentCode  string
	DistrictCode    string
	WorkdayDuration int
	Email           string
	IsInvalid       bool
	// Password — новый пароль; пусто — не менять.
	Password string
	// Caller — кто изменяет: администратор или руководитель теротдела сотрудника.
	Caller auth.Caller
}
