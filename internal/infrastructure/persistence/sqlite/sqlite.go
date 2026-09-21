package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

const userColumns = `id, status, last_name, first_name, middle_name,
	birth_date, position, hired_at, department,
	email, is_invalid, password, created_at, updated_at,
	district, workday_duration`

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

var _ user.UserRepository = (*userRepository)(nil)

func (ur *userRepository) Save(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is nil")
	}

	const query = `
		INSERT INTO users (
			id, status, last_name, first_name, middle_name,
			birth_date, position, hired_at, department,
			email, is_invalid, password,
			district, workday_duration
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
			department       = EXCLUDED.department,
			email            = EXCLUDED.email,
			is_invalid       = EXCLUDED.is_invalid,
			password         = EXCLUDED.password,
			district         = EXCLUDED.district,
			workday_duration = EXCLUDED.workday_duration,
			updated_at       = datetime('now')
	`

	_, err := ur.db.ExecContext(
		ctx,
		query,
		u.ID().String(),
		u.Status().String(),
		u.LastName(),
		u.FirstName(),
		u.MiddleName(),
		u.BirthDate().String(),
		u.Position().String(),
		u.HiredAt().String(),
		u.Department().String(),
		u.Email().Value(),
		boolToInt(u.IsInvalid()),
		u.Password().String(),
		u.District().String(),
		u.WorkdayDuration().Int(),
	)
	if err != nil {
		return fmt.Errorf("saving user: %w", err)
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (ur *userRepository) FindByID(ctx context.Context, userID user.UserID) (*user.User, error) {

	query := `SELECT ` + userColumns + ` FROM users WHERE id = ?`

	var row userRow
	err := ur.db.QueryRowContext(ctx, query, userID.String()).Scan(
		&row.ID, &row.Status,
		&row.LastName, &row.FirstName, &row.MiddleName,
		&row.BirthDate, &row.Position, &row.HiredAt,
		&row.Department,
		&row.Email, &row.IsInvalid, &row.Password,
		&row.CreatedAt, &row.UpdatedAt,
		&row.District, &row.WorkdayDuration,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("finding user by ID: %w", err)
	}

	return row.toDomain()
}

func (ur *userRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	const query = `SELECT ` + userColumns + `FROM users ORDER BY last_name DESC, first_name DESC, middle_name DESC`

	var row userRow
	err := ur.db.QueryRowContext(ctx, query, email.Value()).Scan(
		&row.ID, &row.Status,
		&row.LastName, &row.FirstName, &row.MiddleName,
		&row.BirthDate, &row.Position, &row.HiredAt,
		&row.Department,
		&row.Email, &row.IsInvalid, &row.Password,
		&row.CreatedAt, &row.UpdatedAt,
		&row.District, &row.WorkdayDuration,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}

	return row.toDomain()
}

func (ur *userRepository) FindActive(ctx context.Context) ([]*user.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users ORDER BY last_name DESC, first_name DESC, middle_name DESC`

	rows, err := ur.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying active users: %w", err)
	}
	defer rows.Close()

	users := make([]*user.User, 0, 16)

	for rows.Next() {
		var row userRow
		err := rows.Scan(
			&row.ID, &row.Status,
			&row.LastName, &row.FirstName, &row.MiddleName,
			&row.BirthDate, &row.Position, &row.HiredAt,
			&row.Department,
			&row.Email, &row.IsInvalid, &row.Password,
			&row.CreatedAt, &row.UpdatedAt,
			&row.District, &row.WorkdayDuration,
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

func (ur *userRepository) Delete(ctx context.Context, userID user.UserID) error {
	const query = `
		UPDATE users
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL
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

func (ur *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = ?) AS email_exists;
	`

	var exist bool
	err := ur.db.QueryRowContext(ctx, query, email).Scan(&exist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("checking email: %w", err)
	}

	return exist, nil
}

func (ur *userRepository) FindAll(ctx context.Context) ([]*user.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users ORDER BY last_name DESC, first_name DESC, middle_name DESC`

	rows, err := ur.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	users := make([]*user.User, 0, 16)

	for rows.Next() {
		var row userRow
		err := rows.Scan(
			&row.ID, &row.Status,
			&row.LastName, &row.FirstName, &row.MiddleName,
			&row.BirthDate, &row.Position, &row.HiredAt,
			&row.Department,
			&row.Email, &row.IsInvalid, &row.Password,
			&row.CreatedAt, &row.UpdatedAt,
			&row.District, &row.WorkdayDuration,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}

		u, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("converting user row: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating users: %w", err)
	}

	return users, nil
}
