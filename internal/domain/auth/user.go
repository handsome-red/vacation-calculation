package auth

import "github.com/handsome-red/vacation-calculation/internal/domain/user"

// Caller — кто выполняет действие: администратор, руководитель своих
// территориальных отделов (Managed) или обычный сотрудник.
type Caller struct {
	UserID  user.UserID
	IsAdmin bool
	Managed []string
}

// CallerOf — права пользователя u.
func CallerOf(u *user.User) Caller {
	return Caller{UserID: u.ID(), IsAdmin: u.IsAdmin(), Managed: u.ManagedDistricts()}
}

// Manages — может ли вызывающий действовать как администратор в отношении target.
func (c Caller) Manages(target *user.User) bool {
	return user.Manages(c.IsAdmin, c.Managed, target)
}

// ManagesDistrict — может ли вызывающий управлять территориальным отделом code.
func (c Caller) ManagesDistrict(code string) bool {
	return user.ManagesDistrict(c.IsAdmin, c.Managed, code)
}

// IsStaff — администратор или руководитель.
func (c Caller) IsStaff() bool {
	return c.IsAdmin || len(c.Managed) > 0
}

// Scope — доступные территориальные отделы; nil — все.
func (c Caller) Scope() []string {
	if c.IsAdmin {
		return nil
	}
	if c.Managed == nil {
		return []string{}
	}
	return c.Managed
}

type AuthUser struct {
	id           UserID
	email        Email
	passwordHash PasswordHash
	role         Role
	status       Status
}

func NewAuthUser(id UserID, email Email, passwordHash PasswordHash, role Role, status Status) (*AuthUser, error) {
	if id.IsZero() {
		return nil, ErrIDRequired
	}
	if email.Value() == "" {
		return nil, ErrEmailRequired
	}
	if passwordHash.IsEmpty() {
		return nil, ErrPasswordRequired
	}
	if !role.IsValid() {
		return nil, ErrRoleInvalid
	}
	if !status.IsValid() {
		return nil, ErrStatusInvalid
	}

	return &AuthUser{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		status:       status,
	}, nil
}

func ReconstructAuthUser(
	id UserID,
	email Email,
	passwordHash PasswordHash,
	role Role,
	status Status,
) *AuthUser {
	return &AuthUser{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		status:       status,
	}
}

func (a *AuthUser) ID() UserID                 { return a.id }
func (a *AuthUser) Email() Email               { return a.email }
func (a *AuthUser) PasswordHash() PasswordHash { return a.passwordHash }
func (a *AuthUser) Role() Role                 { return a.role }
func (a *AuthUser) Status() Status             { return a.status }
