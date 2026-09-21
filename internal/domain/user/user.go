package user

import (
	"errors"
	"fmt"
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
	department      Department      // Территориальный отдел
	district        District        // Район/Отдел
	workdayDuration WorkdayDuration // Продолжительность рабочего дня
	email           Email           // Почта
	isInvalid       bool            // Инвалидность
	// totalExperience Experience      // Общий стаж выслуги
	password  Password
	createdAt time.Time
	updatedAt time.Time
}

func (u *User) IsInvalid() bool {
	return u.isInvalid
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

func NewUser(
	id UserID,
	status Status,
	lastName string,
	firstName string,
	middleName string,
	birthDate BirthDate,
	position Position,
	hiredAt HiredDate,
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
		status:          status,
		lastName:        lastName,
		firstName:       firstName,
		middleName:      middleName,
		birthDate:       birthDate,
		position:        position,
		hiredAt:         hiredAt,
		department:      department,
		district:        district,
		workdayDuration: workdayDuration,
		email:           email,
		isInvalid:       isInvalid,
		password:        password,
		createdAt:       now,
		updatedAt:       now,
	}, nil
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

// Deactivate деактивирует аккаунт.
// func (u *User) Deactivate() error {
// 	if !u.isActive {
// 		return errors.New("user is already deactivate")
// 	}

// 	u.isActive = false
// 	u.updatedAt = time.Now().UTC()

// 	return nil
// }

// Activate aктивирует аккаунт.
// func (u *User) Activate() error {
// 	if u.isActive {
// 		return errors.New("user is already activate")
// 	}

// 	u.isActive = true
// 	u.updatedAt = time.Now().UTC()

// 	return nil
// }

func ReconstructUser(
	id UserID,
	email Email,
	password Password,
	firstName string,
	lastName string,
	middleName string,
	department Department,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		id:         id,
		email:      email,
		password:   password,
		firstName:  firstName,
		lastName:   lastName,
		middleName: middleName,
		department: department,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

// func (u *User) Name() Name {
// 	return u.name
// }

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
