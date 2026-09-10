package user

const (
	minimumLength = 8
	maximumLength = 72
)

type Password struct {
	hashed bool
	value  string
}

// NewPasswordFromHash создает Password из существующего хеша.
func NewPasswordFromHash(hash string) Password {
	return Password{
		value:  hash,
		hashed: true,
	}
}

// String возвращает пароль.
func (p Password) String() string {
	return p.value
}

// Hash хэширован ли пароль.
func (p Password) IsHashed() bool {
	return p.hashed
}

// Hash возвращает хеш, если пароль уже захеширован
func (p Password) Hash() (string, error) {
	if !p.hashed {
		return "", ErrPasswordHashing
	}
	return p.value, nil
}

// Валидуируем пароль
func validatePasswordStrenght(password string) error {
	if len(password) < minimumLength {
		return ErrPasswordTooShort
	}

	if len(password) > maximumLength {
		return ErrPasswordTooShort
	}

	return nil
}
