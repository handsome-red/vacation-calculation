package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ user.UserRepository = (*UserRepository)(nil)

func (ur *UserRepository) Save(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is nil")
	}

	const query = `
		INSERT INTO users (id, email, password, first_name, last_name, middle_name, department, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			email      = EXCLUDED.email,
            password   = EXCLUDED.password,
            first_name = EXCLUDED.first_name,
            last_name  = EXCLUDED.last_name,
			middle_name = EXCLUDED.middle_name,
			department = EXCLUDED.department,
			is_acitve = EXCLUDED.is_active
	`

	_, err := ur.db.ExecContext(
		ctx,
		query,
		u.ID().String(),
		u.Email().Value(),
		u.Password().String(),
		u.Name().FirstName(),
		u.Name().LastName(),
		u.Name().MiddleName(),
		u.Department().String(),
		u.IsActive(),
	)
	if err != nil {
		return fmt.Errorf("saving user: %w", err)
	}

	return nil
}

func (ur *UserRepository) FindByID(ctx context.Context, userID user.UserID) (*user.User, error) {
	const query = `
		SELECT
			id, email, password,
			first_name, last_name, middle_name,
			department, is_active,
			created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var row userRow
	err := ur.db.QueryRowContext(ctx, query, userID.String()).Scan(
		&row.ID,
		&row.Email,
		&row.Password,
		&row.FirstName,
		&row.LastName,
		&row.MiddleName,
		&row.Department,
		&row.IsActive,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		// sql.ErrNoRows — это НЕ ошибка. Пользователь просто не найден.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("finding user by ID: %w", err)
	}

	return row.toDomain()
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	const query = `
		SELECT
			id, email, password,
			first_name, last_name, middle_name,
			department, is_active,
			created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var row userRow
	err := ur.db.QueryRowContext(ctx, query, email.Value()).Scan(
		&row.ID,
		&row.Email,
		&row.Password,
		&row.FirstName,
		&row.LastName,
		&row.MiddleName,
		&row.Department,
		&row.IsActive,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}

	return row.toDomain()
}

func (ur *UserRepository) FindActive(ctx context.Context) ([]*user.User, error) {
	const query = `
		SELECT
			id, email, password,
			first_name, last_name, middle_name,
			department, is_active,
			created_at, updated_at
		FROM users
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := ur.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying active users: %w", err)
	}
	defer rows.Close()

	users := make([]*user.User, 0, 16)

	for rows.Next() {
		var row userRow
		err := rows.Scan(
			&row.ID,
			&row.Email,
			&row.Password,
			&row.FirstName,
			&row.LastName,
			&row.MiddleName,
			&row.Department,
			&row.IsActive,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}

		u, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("converting user row to domain: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user rows: %w", err)
	}

	return users, nil
}

func (ur *UserRepository) Delete(ctx context.Context, userID user.UserID) error {
	const query = `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := ur.db.ExecContext(ctx, query, userID.String())
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}
