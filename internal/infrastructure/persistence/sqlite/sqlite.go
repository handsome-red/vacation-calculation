package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

var _ user.UserRepository = (*userRepository)(nil)

const userColumns = `
	u.id, u.status, u.last_name, u.first_name, u.middle_name,
	u.birth_date, u.position, u.hired_at,
	u.department_code, dep.title AS department_title,
	u.district_code,   dis.title AS district_title,
	u.workday_duration, u.email, u.is_invalid, u.password,
	u.created_at, u.updated_at
`

const userFrom = `
	FROM users u
	JOIN departments dep ON dep.code = u.department_code
	JOIN districts   dis ON dis.code = u.district_code
`

func (ur *userRepository) Save(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is nil")
	}

	const query = `
		INSERT INTO users (
			id, status, last_name, first_name, middle_name,
			birth_date, position, hired_at,
			department_code, district_code,
			email, is_invalid, password, workday_duration
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			status           = EXCLUDED.status,
			last_name        = EXCLUDED.last_name,
			first_name       = EXCLUDED.first_name,
			middle_name      = EXCLUDED.middle_name,
			birth_date       = EXCLUDED.birth_date,
			position         = EXCLUDED.position,
			hired_at         = EXCLUDED.hired_at,
			department_code  = EXCLUDED.department_code,
			district_code    = EXCLUDED.district_code,
			email            = EXCLUDED.email,
			is_invalid       = EXCLUDED.is_invalid,
			password         = EXCLUDED.password,
			workday_duration = EXCLUDED.workday_duration,
			updated_at       = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
	`

	_, err := ur.db.ExecContext(
		ctx, query,
		u.ID().String(),
		u.Status().String(),
		u.LastName(),
		u.FirstName(),
		u.MiddleName(),
		u.BirthDate().String(),
		u.Position().String(),
		u.HiredAt().String(),
		u.Department().String(),
		u.District().Code(),
		u.Email().Value(),
		boolToInt(u.IsInvalid()),
		u.Password().String(),
		u.WorkdayDuration().Int(),
	)
	if err != nil {
		return fmt.Errorf("saving user: %w", err)
	}
	return nil
}

func (ur *userRepository) FindByID(ctx context.Context, userID user.UserID) (*user.User, error) {
	var row userRow
	query := `SELECT ` + userColumns + userFrom + ` WHERE u.id = ?`

	if err := ur.db.GetContext(ctx, &row, query, userID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("finding user by ID: %w", err)
	}
	return row.toDomain()
}

func (ur *userRepository) FindAll(ctx context.Context) ([]*user.User, error) {
	var rows []userRow
	query := `SELECT ` + userColumns + userFrom + ` ORDER BY u.last_name, u.first_name, u.middle_name`

	if err := ur.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}

	users := make([]*user.User, 0, len(rows))
	for _, row := range rows {
		u, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("converting user row %q: %w", row.ID, err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (ur *userRepository) Delete(ctx context.Context, userID user.UserID) error {
	const query = `DELETE FROM users WHERE id = ?`

	result, err := ur.db.ExecContext(ctx, query, userID.String())
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return user.ErrUserNotFound
	}
	return nil
}

func (ur *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER(?))`

	var exists bool
	if err := ur.db.QueryRowContext(ctx, query, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking email: %w", err)
	}
	return exists, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
