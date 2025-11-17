package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host, Port, User, Pass, Name string
}

type PostgresRepo struct {
	DB *sql.DB
}

func NewPostgresRepo(cfg *DBConfig) (*PostgresRepo, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresRepo{DB: db}, nil
}

// ==== ADMIN ====

type Admin struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

func (r *PostgresRepo) GetAdminByUsername(ctx context.Context, username string) (*Admin, error) {
	var a Admin
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, username, password_hash, created_at
		 FROM admins WHERE username=$1`, username).
		Scan(&a.ID, &a.Username, &a.PasswordHash, &a.CreatedAt)

	return &a, err
}

// ==== EMPLOYEE ====

type Employee struct {
	ID        string
	Fullname  string
	Email     sql.NullString
	ClickUpID sql.NullString
	Position  sql.NullString
	CreatedAt time.Time
}

func (r *PostgresRepo) ListEmployees(ctx context.Context) ([]Employee, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, fullname, email, clickup_user_id, position, created_at
		 FROM employees ORDER BY fullname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Employee
	for rows.Next() {
		var e Employee
		rows.Scan(&e.ID, &e.Fullname, &e.Email, &e.ClickUpID, &e.Position, &e.CreatedAt)
		out = append(out, e)
	}
	return out, nil
}

func (r *PostgresRepo) GetEmployee(ctx context.Context, id string) (*Employee, error) {
	var e Employee
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, fullname, email, clickup_user_id, position, created_at
		 FROM employees WHERE id=$1`, id).
		Scan(&e.ID, &e.Fullname, &e.Email, &e.ClickUpID, &e.Position, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// ==== TASKS ====

type Task struct {
	ID                  string
	Name                string
	EmployeeID          sql.NullString
	ProjectID           sql.NullString
	Status              sql.NullString
	TimeEstimateSeconds sql.NullInt64
	TimeSpentSeconds    sql.NullInt64
	PercentComplete     sql.NullFloat64
	StartDate           sql.NullTime
	DueDate             sql.NullTime
}

func (r *PostgresRepo) ListTasksByEmployee(ctx context.Context, employeeID string, from, to *time.Time) ([]Task, error) {
	query := `
		SELECT id, name, employee_id, project_id, status,
			   time_estimate_seconds, time_spent_seconds,
			   percent_complete, start_date, due_date
		FROM tasks WHERE employee_id = $1
	`
	args := []interface{}{employeeID}

	if from != nil {
		query += " AND start_date >= $2"
		args = append(args, *from)
	}
	if to != nil {
		query += fmt.Sprintf(" AND start_date <= $%d", len(args)+1)
		args = append(args, *to)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		rows.Scan(
			&t.ID, &t.Name, &t.EmployeeID, &t.ProjectID, &t.Status,
			&t.TimeEstimateSeconds, &t.TimeSpentSeconds,
			&t.PercentComplete, &t.StartDate, &t.DueDate,
		)
		out = append(out, t)
	}
	return out, nil
}

func (r *PostgresRepo) UpsertTask(ctx context.Context, t Task) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO tasks (id, name, employee_id, project_id, status,
			time_estimate_seconds, time_spent_seconds, percent_complete,
			start_date, due_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			name=$2, employee_id=$3, project_id=$4, status=$5,
			time_estimate_seconds=$6, time_spent_seconds=$7,
			percent_complete=$8, start_date=$9, due_date=$10
	`,
		t.ID, t.Name, t.EmployeeID, t.ProjectID, t.Status,
		t.TimeEstimateSeconds, t.TimeSpentSeconds, t.PercentComplete,
		t.StartDate, t.DueDate,
	)
	return err
}
