package auth



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

func (a *AuthUser) ID() UserID              { return a.id }
func (a *AuthUser) Email() Email            { return a.email }
func (a *AuthUser) PasswordHash() PasswordHash { return a.passwordHash }
func (a *AuthUser) Role() Role              { return a.role }
func (a *AuthUser) Status() Status          { return a.status }