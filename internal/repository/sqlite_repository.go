// internal/repository/sqlite_repository.go
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"vacation-calculation/internal/model"
)

type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository - создает новое подключение к SQLite
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Открываем соединение с БД
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Создаем таблицу, если не существует
	if err := createTable(db); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// createTable - создает таблицу vacations
func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS vacations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		employee TEXT NOT NULL,
		start_date DATETIME NOT NULL,
		end_date DATETIME NOT NULL,
		days INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'planned',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (days > 0 AND days <= 365),
		CHECK (status IN ('planned', 'approved', 'completed'))
	);

	CREATE INDEX IF NOT EXISTS idx_vacations_employee ON vacations(employee);
	CREATE INDEX IF NOT EXISTS idx_vacations_status ON vacations(status);
	CREATE INDEX IF NOT EXISTS idx_vacations_start_date ON vacations(start_date);
	`

	_, err := db.Exec(query)
	return err
}

// FindAll - возвращает все отпуска
func (r *SQLiteRepository) FindAll() ([]model.Vacation, error) {
	query := `
	SELECT id, employee, start_date, end_date, days, status, created_at, updated_at
	FROM vacations
	ORDER BY start_date DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query vacations: %w", err)
	}
	defer rows.Close()

	var vacations []model.Vacation
	for rows.Next() {
		var v model.Vacation
		if err := rows.Scan(
			&v.ID,
			&v.Employee,
			&v.StartDate,
			&v.EndDate,
			&v.Days,
			&v.Status,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan vacation: %w", err)
		}
		vacations = append(vacations, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return vacations, nil
}

// FindByID - находит отпуск по ID
func (r *SQLiteRepository) FindByID(id int) (*model.Vacation, error) {
	query := `
	SELECT id, employee, start_date, end_date, days, status, created_at, updated_at
	FROM vacations
	WHERE id = ?
	`

	var v model.Vacation
	err := r.db.QueryRow(query, id).Scan(
		&v.ID,
		&v.Employee,
		&v.StartDate,
		&v.EndDate,
		&v.Days,
		&v.Status,
		&v.CreatedAt,
		&v.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("vacation with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to query vacation: %w", err)
	}

	return &v, nil
}

// Create - создает новый отпуск
func (r *SQLiteRepository) Create(vacation *model.Vacation) error {
	query := `
	INSERT INTO vacations (employee, start_date, end_date, days, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if vacation.CreatedAt.IsZero() {
		vacation.CreatedAt = now
	}
	if vacation.UpdatedAt.IsZero() {
		vacation.UpdatedAt = now
	}
	if vacation.Status == "" {
		vacation.Status = "planned"
	}

	result, err := r.db.Exec(
		query,
		vacation.Employee,
		vacation.StartDate,
		vacation.EndDate,
		vacation.Days,
		vacation.Status,
		vacation.CreatedAt,
		vacation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert vacation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	vacation.ID = int(id)
	return nil
}

// Update - обновляет отпуск
func (r *SQLiteRepository) Update(vacation *model.Vacation) error {
	// Проверяем существование
	_, err := r.FindByID(vacation.ID)
	if err != nil {
		return err
	}

	query := `
	UPDATE vacations
	SET employee = ?, start_date = ?, end_date = ?, days = ?, status = ?, updated_at = ?
	WHERE id = ?
	`

	vacation.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		vacation.Employee,
		vacation.StartDate,
		vacation.EndDate,
		vacation.Days,
		vacation.Status,
		vacation.UpdatedAt,
		vacation.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update vacation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vacation with id %d not found", vacation.ID)
	}

	return nil
}

// Delete - удаляет отпуск
func (r *SQLiteRepository) Delete(id int) error {
	// Проверяем существование
	_, err := r.FindByID(id)
	if err != nil {
		return err
	}

	query := `DELETE FROM vacations WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete vacation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vacation with id %d not found", id)
	}

	return nil
}

// FindByEmployee - находит отпуска по сотруднику
func (r *SQLiteRepository) FindByEmployee(employee string) ([]model.Vacation, error) {
	query := `
	SELECT id, employee, start_date, end_date, days, status, created_at, updated_at
	FROM vacations
	WHERE employee LIKE ?
	ORDER BY start_date DESC
	`

	rows, err := r.db.Query(query, "%"+employee+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query vacations by employee: %w", err)
	}
	defer rows.Close()

	var vacations []model.Vacation
	for rows.Next() {
		var v model.Vacation
		if err := rows.Scan(
			&v.ID,
			&v.Employee,
			&v.StartDate,
			&v.EndDate,
			&v.Days,
			&v.Status,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan vacation: %w", err)
		}
		vacations = append(vacations, v)
	}

	return vacations, nil
}

// FindByStatus - находит отпуска по статусу
func (r *SQLiteRepository) FindByStatus(status string) ([]model.Vacation, error) {
	query := `
	SELECT id, employee, start_date, end_date, days, status, created_at, updated_at
	FROM vacations
	WHERE status = ?
	ORDER BY start_date DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query vacations by status: %w", err)
	}
	defer rows.Close()

	var vacations []model.Vacation
	for rows.Next() {
		var v model.Vacation
		if err := rows.Scan(
			&v.ID,
			&v.Employee,
			&v.StartDate,
			&v.EndDate,
			&v.Days,
			&v.Status,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan vacation: %w", err)
		}
		vacations = append(vacations, v)
	}

	return vacations, nil
}

// GetStats - возвращает статистику по отпускам
func (r *SQLiteRepository) GetStats() (map[string]int, error) {
	query := `
	SELECT status, COUNT(*) as count
	FROM vacations
	GROUP BY status
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer rows.Close()

	stats := map[string]int{
		"total":    0,
		"planned":  0,
		"approved": 0,
		"completed": 0,
	}

	var total int
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats[status] = count
		total += count
	}
	stats["total"] = total

	return stats, nil
}

// Close - закрывает соединение с БД
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}