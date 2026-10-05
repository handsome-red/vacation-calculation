package sqlite

import (
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
)

type authRow struct {
    ID           string `db:"id"`
    Email        string `db:"email"`
    PasswordHash string `db:"password_hash"`
    Role         string `db:"role"`
    Status       string `db:"status"`
}

func (row authRow) toDomain() (*auth.AuthUser, error) {
    id, err := auth.ParseUserID(row.ID)
    if err != nil {
        return nil, fmt.Errorf("auth repo: parse id: %w", err)
    }
    email, err := auth.NewEmail(row.Email)
    if err != nil {
        return nil, fmt.Errorf("auth repo: parse email: %w", err)
    }
    hash := auth.NewPasswordHashUnsafe(row.PasswordHash)
    role, err := auth.NewRole(row.Role)
    if err != nil {
        return nil, fmt.Errorf("auth repo: parse role: %w", err)
    }
    status, err := auth.NewStatus(row.Status)
    if err != nil {
        return nil, fmt.Errorf("auth repo: parse status: %w", err)
    }

    return auth.ReconstructAuthUser(id, email, hash, role, status), nil
}