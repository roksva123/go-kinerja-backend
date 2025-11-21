package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/roksva123/go-kinerja-backend/internal/model"
)

type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

type PostgresRepo struct {
	DB *sql.DB
}

type Repository interface {
    UpsertTeam(ctx context.Context, id string, name string) error
}

func NewPostgresRepoFromConfig(cfg *DBConfig) (*PostgresRepo, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &PostgresRepo{DB: db}, nil
}

func NewPostgresRepo() *PostgresRepo {
	 dsn := os.Getenv("DATABASE_URL")
	 fmt.Println("dsn", dsn)
	 if dsn == "" {
	 dsn = "host=db.fsufakerljrkzrlrjiwm.supabase.co port=5432 user=postgres password=aufa dbname=kinerja_db sslmode=disable"
	 }

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	// verify connection
	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &PostgresRepo{
		DB: db,
	}
}

func (r *PostgresRepo) RunMigrations(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS admins (
         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
         username VARCHAR(100) UNIQUE NOT NULL,
         password_hash TEXT NOT NULL,
         created_at TIMESTAMP DEFAULT NOW()
       );`,
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,
		// users table: stores clickup members
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			clickup_id VARCHAR(50) UNIQUE,
			username VARCHAR(255),
			name VARCHAR(255),
			email VARCHAR(255),
			password TEXT,
			role VARCHAR(50),
			color VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
		);`,
		// teams table: store spaces as teams
		`CREATE TABLE IF NOT EXISTS teams (
			id SERIAL PRIMARY KEY,
			team_id VARCHAR(80) UNIQUE NOT NULL,
			name TEXT NOT NULL,
			parent_id VARCHAR(80),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
		);`,
		// tasks table
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status VARCHAR(100),
			assignee_id BIGINT,
			due_date BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
		);`,
	}
	for _, q := range queries {
		if _, err := r.DB.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

// Admin
type Admin struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

func (r *PostgresRepo) GetAdminByUsername(ctx context.Context, username string) (*Admin, error) {
	var a Admin
	err := r.DB.QueryRowContext(ctx, "SELECT id, username, password_hash, created_at FROM admins WHERE username=$1", username).
		Scan(&a.ID, &a.Username, &a.PasswordHash, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *PostgresRepo) UpsertAdmin(ctx context.Context, username, passwordHash string) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO admins (username, password_hash) VALUES ($1,$2)
		ON CONFLICT (username) DO UPDATE SET password_hash = $2
	`, username, passwordHash)
	return err
}

// Employees
type Employee struct {
	ID        string
	Fullname  string
	Email     sql.NullString
	ClickUpID sql.NullString
	Position  sql.NullString
	CreatedAt time.Time
}

func (r *PostgresRepo) ListEmployees(ctx context.Context) ([]Employee, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, fullname, email, clickup_user_id, position, created_at FROM employees ORDER BY fullname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Employee
	for rows.Next() {
		var e Employee
		if err := rows.Scan(&e.ID, &e.Fullname, &e.Email, &e.ClickUpID, &e.Position, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *PostgresRepo) GetEmployee(ctx context.Context, id string) (*Employee, error) {
	var e Employee
	err := r.DB.QueryRowContext(ctx, `SELECT id, fullname, email, clickup_user_id, position, created_at FROM employees WHERE id=$1`, id).
		Scan(&e.ID, &e.Fullname, &e.Email, &e.ClickUpID, &e.Position, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *PostgresRepo) GetEmployeeByClickUpID(ctx context.Context, clickupID string) (*Employee, error) {
	var e Employee
	err := r.DB.QueryRowContext(ctx, `SELECT id, fullname, email, clickup_user_id, position, created_at FROM employees WHERE clickup_user_id=$1`, clickupID).
		Scan(&e.ID, &e.Fullname, &e.Email, &e.ClickUpID, &e.Position, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *PostgresRepo) UpsertEmployeeFromClickUp(ctx context.Context, fullname, email, clickupID string) (string, error) {
	// tries to upsert based on clickup_user_id or email
	var id string
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO employees (fullname, email, clickup_user_id)
		VALUES ($1,$2,$3)
		ON CONFLICT (clickup_user_id) DO UPDATE SET fullname=EXCLUDED.fullname, email=EXCLUDED.email
		RETURNING id
	`, fullname, email, clickupID).Scan(&id)
	if err == nil {
		return id, nil
	}
	// fallback: check by email
	if email != "" {
		err2 := r.DB.QueryRowContext(ctx, `SELECT id FROM employees WHERE email=$1`, email).Scan(&id)
		if err2 == nil {
			_, _ = r.DB.ExecContext(ctx, `UPDATE employees SET clickup_user_id=$1 WHERE id=$2`, clickupID, id)
			return id, nil
		}
	}
	return "", err
}

// Tasks
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
	UpdatedAt           sql.NullTime
}

func (r *PostgresRepo) UpsertTask(ctx context.Context, t Task) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO tasks (id, name, employee_id, project_id, status, time_estimate_seconds, time_spent_seconds, percent_complete, start_date, due_date, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10, now())
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name,
			employee_id=EXCLUDED.employee_id,
			project_id=EXCLUDED.project_id,
			status=EXCLUDED.status,
			time_estimate_seconds=EXCLUDED.time_estimate_seconds,
			time_spent_seconds=EXCLUDED.time_spent_seconds,
			percent_complete=EXCLUDED.percent_complete,
			start_date=EXCLUDED.start_date,
			due_date=EXCLUDED.due_date,
			updated_at = now()
	`, t.ID, t.Name, t.EmployeeID, t.ProjectID, t.Status, t.TimeEstimateSeconds, t.TimeSpentSeconds, t.PercentComplete, t.StartDate, t.DueDate)
	return err
}

func (r *PostgresRepo) ListTasksByEmployee(ctx context.Context, employeeID string, from, to *time.Time) ([]Task, error) {
	q := `SELECT id, name, employee_id, project_id, status, time_estimate_seconds, time_spent_seconds, percent_complete, start_date, due_date FROM tasks WHERE employee_id=$1`
	args := []interface{}{employeeID}
	if from != nil {
		q += " AND (start_date IS NOT NULL AND start_date >= $2)"
		args = append(args, *from)
	}
	if to != nil {
		q += fmt.Sprintf(" AND (start_date IS NOT NULL AND start_date <= $%d)", len(args)+1)
		args = append(args, *to)
	}
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Name, &t.EmployeeID, &t.ProjectID, &t.Status, &t.TimeEstimateSeconds, &t.TimeSpentSeconds, &t.PercentComplete, &t.StartDate, &t.DueDate); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (r *PostgresRepo) UpsertUser(ctx context.Context, u *model.User) error {
	
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO users (clickup_id, username, name, email, password, role, color, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7, now())
		ON CONFLICT (clickup_id) DO UPDATE SET
			username = EXCLUDED.username,
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			password = COALESCE(EXCLUDED.password, users.password),
			role = EXCLUDED.role,
			color = EXCLUDED.color,
			updated_at = now()
	`, u.Username,
		u.Username,
		u.Name,
		u.Email,
		u.Password,
		u.Role,
		u.Color,
	)
	return err
}



func (r *PostgresRepo) GetAllUsers(ctx context.Context) ([]model.User, error) {
    query := `SELECT id, username, name, role, created_at, updated_at FROM users ORDER BY id`

    rows, err := r.DB.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []model.User
    for rows.Next() {
        var u model.User
        if err := rows.Scan(
            &u.ID, &u.Username, &u.Name, &u.Role,
            &u.CreatedAt, &u.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    return users, nil
}



func (r *PostgresRepo) UpsertTeam(ctx context.Context, teamID, name, parentID string) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO teams (team_id, name, parent_id)
		VALUES ($1,$2,$3)
		ON CONFLICT (team_id) DO UPDATE SET
			name = EXCLUDED.name,
			parent_id = EXCLUDED.parent_id,
			updated_at = now()
	`, teamID, name, parentID)
	return err
}





func (r *PostgresRepo) UpsertEmployee(ctx context.Context, emp *model.Employee) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO employees (id, fullname, email, role, team_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(id) DO UPDATE SET
			fullname = EXCLUDED.name,
			email = EXCLUDED.email,
			role = EXCLUDED.role,
			team_id = EXCLUDED.team_id,
			updated_at = NOW()
	`,
		emp.ID,
		emp.Name,
		emp.Email,
		emp.Role,
		emp.ID,
	)

	return err
}
func (r *PostgresRepo) GetTeams(ctx context.Context) ([]model.Team, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT team_id, name, parent_id, created_at, updated_at FROM teams ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Team
	for rows.Next() {
		var t model.Team
		var parent sql.NullString
		if err := rows.Scan(&t.ID, &t.Name, &parent, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if parent.Valid {
			p := parent.String
			t.ParentID = &p
		} else {
			t.ParentID = nil
		}
		out = append(out, t)
	}
	return out, nil
}



func (r *PostgresRepo) GetUsers(ctx context.Context) ([]model.User, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, username, name, email, role, color, created_at, updated_at FROM users ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.Role, &u.Color, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}


func (r *PostgresRepo) GetTasks(ctx context.Context) ([]model.Task, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name, status, assignee_id, due_date FROM tasks ORDER BY due_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Name, &t.Status, &t.AssigneeID, &t.DueDate); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
