package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

var _ ports.UserRepository = (*userRepository)(nil)

const userColumns = `
	u.id, u.status, u.last_name, u.first_name, u.middle_name,
	u.birth_date, u.position, u.hired_at, u.experience,
	u.department_code, dep.title AS department_title,
	u.district_code,   dis.title AS district_title,
	u.workday_duration, u.email, u.is_invalid, u.password, u.superadmin,
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
			experience,
			department_code, district_code,
			email, is_invalid, password, workday_duration, superadmin
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			status           = EXCLUDED.status,
			last_name        = EXCLUDED.last_name,
			first_name       = EXCLUDED.first_name,
			middle_name      = EXCLUDED.middle_name,
			birth_date       = EXCLUDED.birth_date,
			position         = EXCLUDED.position,
			hired_at         = EXCLUDED.hired_at,
			experience       = EXCLUDED.experience,
			department_code  = EXCLUDED.department_code,
			district_code    = EXCLUDED.district_code,
			email            = EXCLUDED.email,
			is_invalid       = EXCLUDED.is_invalid,
			password         = EXCLUDED.password,
			workday_duration = EXCLUDED.workday_duration,
			superadmin       = EXCLUDED.superadmin,
			updated_at       = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
	`

	_, err := ur.db.ExecContext(
		ctx, query,
		u.ID().String(),
		u.Status().String(),
		u.LastName(),
		u.FirstName(),
		u.MiddleName(),
		u.BirthDate().ISO(),
		u.Position().String(),
		u.HiredAt().Time().Format("2006-01-02"),
		u.ExperienceDays(),
		u.Department().Code,
		u.District().Code(),
		u.Email().Value(),
		boolToInt(u.IsInvalid()),
		u.Password().String(),
		u.WorkdayDuration().Int(),
		boolToInt(u.IsAdmin()),
	)
	if err != nil {
		return fmt.Errorf("saving user: %w", err)
	}
	return nil
}

func (ur *userRepository) ExistsAdmin(ctx context.Context) (bool, error) {
	var exists bool
	if err := ur.db.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM users WHERE superadmin = 1)`); err != nil {
		return false, fmt.Errorf("check admin exists: %w", err)
	}
	return exists, nil
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

func (ur *userRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	var row userRow
	query := `SELECT ` + userColumns + userFrom + ` WHERE LOWER(u.email) = LOWER(?)`
	if err := ur.db.GetContext(ctx, &row, query, email.Value()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return row.toDomain()
}

func (ur *userRepository) FindUsers(ctx context.Context, q ports.Query) ([]*user.User, int, error) {
	where, args := buildWhere(q.Filters)

	var total int
	countSQL := "SELECT COUNT(*)" + userFrom + where
	if err := ur.db.GetContext(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	dataArgs := append([]any{}, args...)
	dataArgs = append(dataArgs, q.Size, (q.Page-1)*q.Size)

	dataSQL := "SELECT " + userColumns + userFrom + where + `
        ORDER BY u.last_name, u.first_name, u.middle_name
        LIMIT ? OFFSET ?
    `

	var rows []userRow
	if err := ur.db.SelectContext(ctx, &rows, dataSQL, dataArgs...); err != nil {
		return nil, 0, fmt.Errorf("querying users: %w", err)
	}

	users := make([]*user.User, 0, len(rows))
	for _, r := range rows {
		u, err := r.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("converting user row %q: %w", r.ID, err)
		}
		users = append(users, u)
	}

	return users, total, nil
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

func buildWhere(f ports.UserFilter) (string, []any) {
	var (
		conditions []string
		args       []any
	)

	add := func(cond string, val any) {
		conditions = append(conditions, cond)
		args = append(args, val)
	}

	// LIKE и LOWER в SQLite не учитывают регистр только для латиницы, поэтому
	// для кириллицы сравниваем с вариантами: как введено, строчными, с заглавной.
	addLike := func(column, value string) {
		variants := caseVariants(value)
		parts := make([]string, len(variants))
		for i, v := range variants {
			parts[i] = column + " LIKE ?"
			args = append(args, "%"+v+"%")
		}
		conditions = append(conditions, "("+strings.Join(parts, " OR ")+")")
	}

	if f.Status != "" {
		add("u.status = ?", f.Status)
	}
	if f.LastName != "" {
		addLike("u.last_name", f.LastName)
	}
	if f.FirstName != "" {
		addLike("u.first_name", f.FirstName)
	}
	if f.MiddleName != "" {
		addLike("u.middle_name", f.MiddleName)
	}
	if f.Email != "" {
		addLike("u.email", f.Email)
	}
	if f.DistrictCode != "" {
		add("u.district_code = ?", f.DistrictCode)
	}
	if f.DepartmentCode != "" {
		add("u.department_code = ?", f.DepartmentCode)
	}
	if f.Position != "" {
		add("u.position = ?", f.Position)
	}
	if len(f.Positions) > 0 {
		marks := strings.TrimSuffix(strings.Repeat("?,", len(f.Positions)), ",")
		conditions = append(conditions, "u.position IN ("+marks+")")
		for _, p := range f.Positions {
			args = append(args, p)
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

// caseVariants — строка как введена, строчными и с заглавной первой буквой.
func caseVariants(s string) []string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	r := []rune(lower)
	title := strings.ToUpper(string(r[:1])) + string(r[1:])

	out := []string{s}
	for _, v := range []string{lower, title} {
		if v != out[0] && (len(out) < 2 || v != out[1]) {
			out = append(out, v)
		}
	}
	return out
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
