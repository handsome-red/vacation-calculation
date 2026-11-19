package user

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	id              UserID          // ID
	status          Status          // Статус
	lastName        string          // Фамилия
	firstName       string          // Имя
	middleName      string          // Отчество
	birthDate       BirthDate       // Дата рождения
	position        Position        // Должность
	hiredAt         HiredDate       // Дата приема на работу
	experienceDays  int				// Дней стажа до приема на работу
	department      Department      // Территориальный отдел
	district        District        // Район/Отдел
	workdayDuration WorkdayDuration // Продолжительность рабочего дня
	email           Email           // Почта
	isInvalid       bool            // Инвалидность
	superadmin      bool            // Суперадминистратор
	password        Password
	createdAt       time.Time
	updatedAt       time.Time
}

type UserParams struct {
	ID              UserID
	Status          Status
	Email           Email
	Password        Password
	FirstName       string
	LastName        string
	MiddleName      string
	BirthDate       BirthDate
	Position        Position
	HiredAt         HiredDate
	ExperienceDays  int
	Department      Department
	District        District
	WorkdayDuration WorkdayDuration
	IsInvalid       bool
	Superadmin      bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewUser(
	id UserID,
	lastName string,
	firstName string,
	middleName string,
	birthDate BirthDate,
	position Position,
	hiredAt HiredDate,
	experienceDays int,
	department Department,
	district District,
	workdayDuration WorkdayDuration,
	email Email,
	isInvalid bool,
	password Password,
) (*User, error) {

	now := time.Now().UTC()

	return &User{
		id:              id,
		lastName:        lastName,
		firstName:       firstName,
		middleName:      middleName,
		birthDate:       birthDate,
		position:        position,
		hiredAt:         hiredAt,
		experienceDays:  experienceDays,
		department:      department,
		district:        district,
		workdayDuration: workdayDuration,
		email:           email,
		isInvalid:       isInvalid,
		password:        password,
		createdAt:       now,
		updatedAt:       now,
		status:          StatusActive,
	}, nil
}

func ReconstructUser(p UserParams) *User {
	return &User{
		id:              p.ID,
		status:          p.Status,
		email:           p.Email,
		password:        p.Password,
		firstName:       p.FirstName,
		lastName:        p.LastName,
		middleName:      p.MiddleName,
		birthDate:       p.BirthDate,
		position:        p.Position,
		hiredAt:         p.HiredAt,
		experienceDays:  p.ExperienceDays,
		department:      p.Department,
		district:        p.District,
		workdayDuration: p.WorkdayDuration,
		isInvalid:       p.IsInvalid,
		superadmin:      p.Superadmin,
		createdAt:       p.CreatedAt,
		updatedAt:       p.UpdatedAt,
	}
}

func (u *User) Experience() Experience {
	from := u.hiredAt.Time()
	to := time.Now()

	totalDays := daysBetween(from, to)
	y, m, d := splitExperience(from, to)

	return Experience{
		Years:     y,
		Months:    m,
		Days:      d,
		TotalDays: totalDays,
	}
}

func splitExperience(from, to time.Time) (int, int, int) {
	if to.Before(from) {
		return 0, 0, 0
	}

	years := to.Year() - from.Year()
	if from.AddDate(years, 0, 0).After(to) {
		years--
	}

	base := from.AddDate(years, 0, 0)
	months := 0
	for {
		next := base.AddDate(0, 1, 0)
		if next.After(to) {
			break
		}
		base = next
		months++
	}

	days := daysBetween(base, to)

	return years, months, days
}

func daysBetween(from, to time.Time) int {
	f := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	t := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	return int(t.Sub(f) / (24 * time.Hour))
}

func (u *User) IsActive() bool {
	return u.Status().IsActive()
}

func (u *User) IsIrregular() bool {
	return u.Position().isIrregular()
}

func (u *User) ExperienceLabel() string {
	return u.Experience().String()
}

func (u *User) IsInvalid() bool {
	return u.isInvalid
}

func (u *User) isAdmin() bool {
	return u != nil && u.superadmin
}

// GrantAdmin делает пользователя суперадминистратором.
func (u *User) GrantAdmin() {
	u.superadmin = true
	u.updatedAt = time.Now().UTC()
}

// IsAdmin сообщает, является ли пользователь суперадминистратором.
func (u *User) IsAdmin() bool {
	return u.isAdmin()
}

func (u *User) District() District {
	return u.district
}

func (u *User) WorkdayDuration() WorkdayDuration {
	return u.workdayDuration
}

func (u *User) HiredAt() HiredDate {
	return u.hiredAt
}

func (u *User) Position() Position {
	return u.position
}

func (u *User) BirthDate() BirthDate {
	return u.birthDate
}

func (u *User) Status() Status {
	return u.status
}

func (u *User) FirstName() string {
	return u.firstName
}

func (u *User) LastName() string {
	return u.lastName
}

func (u *User) MiddleName() string {
	return u.middleName
}

func (u *User) FullName() string {
	return fmt.Sprintf("%s %s %s", u.lastName, u.firstName, u.middleName)
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Department() Department {
	return u.department
}

func (u *User) Password() Password {
	return u.password
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) ExperienceDays() int {
    return u.experienceDays
}

// Initials — «Фамилия И.О.»; инициал пропускается, если имени или отчества нет.
func (u *User) Initials() string {
	s := u.LastName()
	if f := []rune(strings.TrimSpace(u.FirstName())); len(f) > 0 {
		s += " " + string(f[0]) + "."
		if m := []rune(strings.TrimSpace(u.MiddleName())); len(m) > 0 {
			s += string(m[0]) + "."
		}
	}
	return s
}

// Profile — данные сотрудника, которые может менять администратор.
type Profile struct {
	LastName        string
	FirstName       string
	MiddleName      string
	BirthDate       BirthDate
	Position        Position
	HiredAt         HiredDate
	ExperienceDays  int
	Department      Department
	District        District
	WorkdayDuration WorkdayDuration
	Email           Email
	IsInvalid       bool
}

// UpdateProfile заменяет данные сотрудника. Пароль, статус и права не меняются.
func (u *User) UpdateProfile(p Profile) error {
	if p.LastName == "" || p.FirstName == "" || p.MiddleName == "" {
		return ErrNameRequired
	}
	if p.ExperienceDays < 0 {
		return ErrExperienceInvalid
	}
	u.lastName = p.LastName
	u.firstName = p.FirstName
	u.middleName = p.MiddleName
	u.birthDate = p.BirthDate
	u.position = p.Position
	u.hiredAt = p.HiredAt
	u.experienceDays = p.ExperienceDays
	u.department = p.Department
	u.district = p.District
	u.workdayDuration = p.WorkdayDuration
	u.email = p.Email
	u.isInvalid = p.IsInvalid
	u.updatedAt = time.Now().UTC()
	return nil
}

// ChangeEmail изменение почты пользователя
func (u *User) ChangeEmail(newEmail Email) error {
	if u.email == newEmail {
		return errors.New("new email must be different from current")
	}

	u.email = newEmail
	u.updatedAt = time.Now().UTC()
	return nil
}

// ChangePassword меняет пароль пользователя.
func (u *User) ChangePassword(newPassword Password) error {
	if u.password == newPassword {
		return errors.New("new password must be different from current")
	}

	u.password = newPassword
	u.updatedAt = time.Now().UTC()

	return nil
}
