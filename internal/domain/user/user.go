package user

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Email struct {
	value string
}

// NewEmail создает новый Email с валидацией.
// Возвращает ошибку, если email некорректен.
func NewEmail(email string) (Email, error) {
	if err := validateEmail(email); err != nil {
		return Email{}, err
	}

	return Email{value: email}, nil
}

// Value возвращает строковое представление Email.
func (e Email) Value() string {
	return e.value
}

// String реализует интерфейс fmt.Stringer
func (e Email) String() string {
	return e.value
}

// validateEmail проверяет корректность email.
func validateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

type Password struct {
	hash string
}

// NewPasswordFromHash создает Password из существующего хеша.
func NewPasswodFromHash(hash string) Password {
	return Password{hash: hash}
}

// Hash возвращает хеш пароля.
func (p Password) Hash() string {
	return p.hash
}

type UserID struct {
	value uuid.UUID
}

// NewUserID создает новый UserID.
func NewUserID(value uuid.UUID) UserID {
	return UserID{value: value}
}

// String возвращает строковое представление ID.
func (id UserID) String() string {
	return id.Value().String()
}

// Value возвращает значение ID.
func (id UserID) Value() uuid.UUID {
	return id.value
}

type User struct {
	id         UserID
	email      Email
	password   Password
	firstName  string
	lastName   string
	middleName string
	department string
	createdAt  time.Time
	updatedAt  time.Time
	isActive   bool
}

func NewUser(
	id UserID,
	email Email,
	password Password,
	firstname, lastname, middlename, department string,
) (*User, error) {
	if firstname == "" || lastname == "" || middlename == "" {
		return nil, errors.New("first name, lastname and middlename cannot be empty")
	}

	now := time.Now().UTC()

	return &User{
		id:         id,
		email:      email,
		password:   password,
		firstName:  firstname,
		lastName:   lastname,
		middleName: middlename,
		department: department,
		createdAt:  now,
		updatedAt:  now,
		isActive:   true,
	}, nil
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
func (u *User) Deactivate() error {
	if !u.isActive {
		return errors.New("user is already deactivate")
	}

	u.isActive = false
	u.updatedAt = time.Now().UTC()

	return nil
}

// Activate aктивирует аккаунт.
func (u *User) Activate() error {
	if u.isActive {
		return errors.New("user is already activate")
	}

	u.isActive = true
	u.updatedAt = time.Now().UTC()

	return nil
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Password() Password {
	return u.password
}

func (u *User) FirstName() string {
	return u.firstName
}

func (u *User) LastName() string {
	return u.lastName
}

func (u *User) FullName() string {
	return fmt.Sprintf("%s %s", u.firstName, u.lastName)
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) IsActive() bool {
	return u.isActive
}
