package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

type Admin struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type PostgresRepo struct {
    DB *sql.DB
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

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &PostgresRepo{
		DB: db,
	}
}

func (r *PostgresRepo) GetByUsername(ctx context.Context, username string) (*model.Admin, error) {
    row := r.DB.QueryRowContext(ctx,
        `SELECT id, username, password_hash, created_at 
         FROM admins WHERE username = $1 LIMIT 1`, username)

    var a model.Admin
    err := row.Scan(&a.ID, &a.Username, &a.PasswordHash, &a.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &a, nil
}


func (r *PostgresRepo) RunMigrations(ctx context.Context) error {
    queries := []string{
		`DROP TABLE IF EXISTS task_assignees;`,
		`DROP TABLE IF EXISTS users CASCADE;`,
		`CREATE TABLE IF NOT EXISTS admins (
         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
         username VARCHAR(100) UNIQUE NOT NULL,
         password_hash TEXT NOT NULL,
         created_at TIMESTAMP DEFAULT NOW()
       );`,
        `CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,
        `CREATE TABLE IF NOT EXISTS teams (
            team_id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            parent_id TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
        );`,
        `CREATE TABLE IF NOT EXISTS users (
         clickup_id BIGINT PRIMARY KEY,
         name TEXT,
         username TEXT,
         email TEXT,
         password TEXT,
         role TEXT,
         status TEXT NOT NULL DEFAULT 'active',
         created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
         updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
        );`,
        `CREATE TABLE IF NOT EXISTS tasks (
        id TEXT PRIMARY KEY,
        name TEXT,
        text_content TEXT,
        description TEXT,
    
        status_id TEXT,
        status_name TEXT,
        status_type TEXT,
        status_color TEXT,
    
        date_done BIGINT,
        date_closed BIGINT,
        start_date BIGINT,
        due_date BIGINT,
        time_estimate BIGINT,
        time_spent_ms BIGINT,
    
        assignee_clickup_id BIGINT,
        assignee_user_id BIGINT,
        assignee_username TEXT,
        assignee_email TEXT,
        assignee_color TEXT,
    
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );`,
    `CREATE TABLE IF NOT EXISTS task_assignees (
        id BIGSERIAL PRIMARY KEY,
        task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
        user_clickup_id BIGINT NOT NULL REFERENCES users(clickup_id) ON DELETE CASCADE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
        UNIQUE(task_id, user_clickup_id)
    );`,
    `CREATE INDEX IF NOT EXISTS idx_tasks_start_date ON tasks(start_date);
     CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
     CREATE INDEX IF NOT EXISTS idx_tasks_status_name ON tasks(status_name);
     CREATE INDEX IF NOT EXISTS idx_tasks_assignee_username ON tasks(assignee_username);`,
    `CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    task_id TEXT,
    name TEXT,
    color TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_tags_task_id ON tags(task_id);`,
    `CREATE TABLE IF NOT EXISTS audit_logs (
     id BIGSERIAL PRIMARY KEY,
     action TEXT,
     user_id BIGINT,
     payload JSONB,
     created_at TIMESTAMPTZ DEFAULT now()
     );`,
    }
    for _, q := range queries {
        if _, err := r.DB.ExecContext(ctx, q); err != nil {
            return err
        }
    }
    return nil
}

func (r *PostgresRepo) GetAdminByUsername(ctx context.Context, username string) (*model.Admin, error) {
    query := `
        SELECT id, username, password_hash
        FROM admins
        WHERE username = $1
        LIMIT 1
    `

    row := r.DB.QueryRowContext(ctx, query, username)
    fmt.Println("DEBUG mencari username:", username)
    

    var a model.Admin
    err := row.Scan(
        &a.ID,
        &a.Username,
        &a.PasswordHash,
    )
    if err != nil {
        fmt.Println("SCAN ERROR:", err)
        return nil, err
    }


    return &a, nil
}



// Get admin by username
func (r *PostgresRepo) GetAllUsers(ctx context.Context) ([]model.User, error) { // Tetap ada, tapi tidak lagi digunakan secara luas
    query := `SELECT clickup_id, username, name, role, created_at, updated_at FROM users ORDER BY username`

    rows, err := r.DB.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []model.User
    for rows.Next() {
        var u model.User
        if err := rows.Scan(
            &u.ClickUpID, &u.Name, &u.Name, &u.Role,
            &u.CreatedAt, &u.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    return users, nil
}

func (r *PostgresRepo) UpsertAdmin(ctx context.Context, username, passwordHash string) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO admins (username, password_hash) VALUES ($1,$2)
		ON CONFLICT (username) DO UPDATE SET password_hash = $2
	`, username, passwordHash)
	return err
}

func (r *PostgresRepo) UpsertTaskAssignees(ctx context.Context, taskID string, assigneeIDs []int64) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback jika ada error

	// 1. Hapus semua assignee lama untuk task ini
	_, err = tx.ExecContext(ctx, "DELETE FROM task_assignees WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete old assignees: %w", err)
	}

	if len(assigneeIDs) == 0 {
		return tx.Commit() // Jika tidak ada assignee baru, cukup commit dan selesai
	}

	// 2. Sisipkan semua assignee baru
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO task_assignees (task_id, user_clickup_id) VALUES ($1, $2) ON CONFLICT DO NOTHING")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, userID := range assigneeIDs {
		if _, err := stmt.ExecContext(ctx, taskID, userID); err != nil {
			return fmt.Errorf("failed to insert assignee %d for task %s: %w", userID, taskID, err)
		}
	}

	return tx.Commit()
}

// UpsertTeam
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

// UpsertUser
func (r *PostgresRepo) UpsertUser(ctx context.Context, u *model.User) error {
	_, err := r.DB.ExecContext(ctx, `
        INSERT INTO users (clickup_id, username, name, email, role, status)
        VALUES ($1, $2, $2, $3, $4, $5)
        ON CONFLICT (clickup_id) DO UPDATE SET
            username = EXCLUDED.username,
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            role = EXCLUDED.role,
            status = EXCLUDED.status,
            updated_at = now()
    `,
		u.ClickUpID,
		u.Name,
		u.Email,
		u.Role,
        u.Status,
	)
	return err
}


// UpsertTask
func (r *PostgresRepo) UpsertTask(ctx context.Context, t *model.TaskResponse) error {
    query := `
        INSERT INTO tasks (
            id, name, text_content, description,
            status_id, status_name, status_type, status_color,
            date_done, date_closed, start_date, due_date,
            assignee_user_id,
            time_estimate, time_spent_ms
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        ON CONFLICT (id)
        DO UPDATE SET
            name = EXCLUDED.name,
            text_content = EXCLUDED.text_content,
            description = EXCLUDED.description,
            status_id = EXCLUDED.status_id,
            status_name = EXCLUDED.status_name,
            status_color = EXCLUDED.status_color,
            date_done = EXCLUDED.date_done,
            date_closed = EXCLUDED.date_closed,
            start_date = EXCLUDED.start_date,
            due_date = EXCLUDED.due_date,
            assignee_user_id = EXCLUDED.assignee_user_id,
            time_estimate = EXCLUDED.time_estimate,
            time_spent_ms = EXCLUDED.time_spent_ms,
            updated_at = now()
    `

    _, err := r.DB.ExecContext(ctx, query,
        t.ID,
        t.Name,
        t.TextContent,
        t.Description,
        t.Status.ID,
        t.Status.Name,
        t.Status.Type,
        t.Status.Color,
        t.DateDone,
        t.DateClosed,
        t.StartDate,
        t.DueDate,
        t.AssigneeUserID,
        t.TimeEstimate,
        t.TimeSpentMs,
    )

    return err
}

func (r *PostgresRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
    query := `
        SELECT clickup_id, username, email, role
        FROM users
        WHERE email = $1
        LIMIT 1
    `

    var u model.User
    err := r.DB.QueryRowContext(ctx, query, email).Scan(&u.ClickUpID,
        &u.Name,
        &u.Email,
        &u.Role,
    )

    if err == sql.ErrNoRows {
        return nil, nil
    }

    if err != nil {
        return nil, err
    }

    return &u, nil
}


func (r *PostgresRepo) GetTasks(ctx context.Context) ([]model.TaskResponse, error) {
    q := `
        SELECT 
            COALESCE(task_id, id) AS id,
            name,
            text_content,
            description,
            status_id,
            status_name,
            status_type,
            status_color,
            date_done,
            start_date,
            due_date,
            date_closed,
            assignee_username,
            assignee_email,
            assignee_color
        FROM tasks
        ORDER BY COALESCE(date_done, 0) DESC
    `

    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    out := []model.TaskResponse{}

    for rows.Next() {

        var t model.TaskResponse

        var (
            statusID, statusName, statusType, statusColor sql.NullString
            dateDone, dateClosed sql.NullInt64
            startDate, dueDate sql.NullInt64
            uname, uemail, ucolor sql.NullString
        )

        if err := rows.Scan(
            &t.ID,
            &t.Name,
            &t.TextContent,
            &t.Description,
            &statusID,
            &statusName,
            &statusType,
            &statusColor,
            &dateDone,
            &dueDate,
            &startDate,
            &dateClosed,
            &uname,
            &uemail,
            &ucolor,
        ); err != nil {
            return nil, err
        }

        t.Status.ID = statusID.String
        t.Status.Name = statusName.String
        t.Status.Type = statusType.String
        t.Status.Color = statusColor.String

        if dateDone.Valid {
            v := dateDone.Int64
            t.DateDone = &v
        }

        if dateClosed.Valid {
            v := dateClosed.Int64
            t.DateClosed = &v
        }

        if dueDate.Valid {
        v := dueDate.Int64
        t.DueDate = &v
        }

        if startDate.Valid {
        v := startDate.Int64
        t.StartDate = &v
        }
        if uname.Valid {
            t.Username = uname.String
        }
        if uemail.Valid {
            t.Email = uemail.String
        }
        if ucolor.Valid {
            t.Color = ucolor.String
        }

        out = append(out, t)
    }

    return out, nil
}





func (r *PostgresRepo) GetMembers(ctx context.Context) ([]model.User, error) { 
    q := `SELECT clickup_id, name, email, role, status, created_at, updated_at FROM users ORDER BY name`
    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []model.User
    for rows.Next() {
        var u model.User
        if err := rows.Scan(&u.ClickUpID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, err
        }
        out = append(out, u)
    }
    return out, nil
}

// GetTeams
func (r *PostgresRepo) GetTeams(ctx context.Context) ([]model.Team, error) {
    q := `SELECT team_id, name, parent_id, created_at, updated_at FROM teams ORDER BY name`
    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []model.Team
    for rows.Next() {
        var t model.Team
        var parent sql.NullString
        if err := rows.Scan(&t.TeamID, &t.Name, &parent, &t.CreatedAt, &t.UpdatedAt); err != nil {
            return nil, err
        }
        if parent.Valid {
            t.ParentID = &parent.String
        }
        out = append(out, t)
    }
    return out, nil
}



// GetFullSyncFiltered 
func (r *PostgresRepo) GetFullSyncFiltered(ctx context.Context, start, end *int64, role string) ([]model.TaskWithMember, error) {

    q := `SELECT 1` // Query ini tampaknya tidak lagi relevan/digunakan

    args := []interface{}{}
    idx := 1

    // filter tanggal
    if start != nil && end != nil {
        q += fmt.Sprintf(`
            AND (
                (t.date_created IS NOT NULL AND t.date_created BETWEEN $%d AND $%d) OR
                (t.date_done IS NOT NULL AND t.date_done BETWEEN $%d AND $%d) OR
                (t.date_closed IS NOT NULL AND t.date_closed BETWEEN $%d AND $%d)
            )
        `, idx, idx+1, idx, idx+1, idx, idx+1)

        args = append(args, *start, *end)
        idx += 2
    }

    // filter role karyawan
    if role != "" {
        args = append(args, role)
        idx++
    }

    q += " ORDER BY t.date_created DESC "

    rows, err := r.DB.QueryContext(ctx, q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.TaskWithMember

    for rows.Next() {
        var t model.TaskWithMember
        var mID sql.NullInt64
        var mU, mE, mC sql.NullString
        

        err := rows.Scan(
            &t.TaskID, &t.TaskName, &t.TaskStatus,
            &t.DateCreated, &t.DateDone, &t.DateClosed,
            &mID, &mU, &mE, &mC,
        )
        if err != nil {
            return nil, err
        }

        if mID.Valid {
            t.UserID = mID.Int64
            t.Username = mU.String
            t.Email = mE.String
            t.Color = mC.String
        }

        out = append(out, t)
    }

    return out, nil
}

func (r *PostgresRepo) GetFullDataFiltered(ctx context.Context, startMs, endMs *int64, role, username string) ([]model.TaskWithMember, error) {
    q := `
        SELECT 
            t.id, t.name, t.description, 
            t.status_name, t.status_type, t.status_color,
            t.start_date, t.due_date, t.date_done, t.date_closed, t.time_estimate, t.time_spent,
            u.clickup_id, u.username, u.email, u.role
        FROM tasks t
        LEFT JOIN users u ON t.assignee_user_id = u.clickup_id
        WHERE 1=1
    `

    args := []interface{}{}
    idx := 1

    if startMs != nil && endMs != nil {
        q += fmt.Sprintf(" AND ( (t.start_date IS NOT NULL AND t.start_date BETWEEN $%d AND $%d) OR (t.due_date IS NOT NULL AND t.due_date BETWEEN $%d AND $%d) )", idx, idx+1, idx, idx+1)
        args = append(args, *startMs, *endMs)
        idx += 2
    }

    if role != "" {
        args = append(args, role)
        idx++
    }

    if username != "" {
        q += fmt.Sprintf(" AND (t.assignee_username ILIKE $%d OR u.username ILIKE $%d)", idx, idx)
        args = append(args, "%"+username+"%", "%"+username+"%")
        idx += 2
    }

    q += " ORDER BY COALESCE(t.start_date, t.date_done, 0) DESC"

    rows, err := r.DB.QueryContext(ctx, q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.TaskWithMember
    for rows.Next() {
        var t model.TaskWithMember
        var userID sql.NullInt64
        var userUsername, userEmail, userRole sql.NullString
        var startDate, dueDate, dateDone, dateClosed, timeEstimate, timeSpent sql.NullInt64

        err := rows.Scan(
            &t.TaskID, &t.TaskName, &t.TaskDescription,
            &t.TaskStatus, &t.TaskStatusType, &t.TaskStatusColor,
            &startDate, &dueDate, &dateDone, &dateClosed, &timeEstimate, &timeSpent,
            &userID, &userUsername, &userEmail, &userRole,
        )
        if err != nil {
            return nil, err
        }

        if startDate.Valid { v := startDate.Int64; t.StartDate = &v }
        if dueDate.Valid { v := dueDate.Int64; t.DueDate = &v }
        if dateDone.Valid { v := dateDone.Int64; t.DateDone = &v }
        if dateClosed.Valid { v := dateClosed.Int64; t.DateClosed = &v }
        if timeEstimate.Valid { v := timeEstimate.Int64; t.TimeEstimate = &v }
        if timeSpent.Valid { v := timeSpent.Int64; t.TimeSpent = &v }

        if userID.Valid {
            t.UserID = userID.Int64
            t.Username = userUsername.String
            t.Email = userEmail.String
            t.Role = userRole.String
        }

        out = append(out, t)
    }

    return out, nil
}

func (r *PostgresRepo) GetTasksByRange(
    ctx context.Context,
    startMs *int64,
    endMs *int64,
    username string,
    status string,
) ([]model.Task, error) {

    query := `
        SELECT 
            id,
            task_id,
            name,
            text_content,
            description,
            status_id,
            status_name,
            status_type,
            status_color,
            date_done,
            date_closed,
            start_date,
            due_date,
            assignee_username,
            assignee_email,
            assignee_color,
            assignee_user_id,
            assignee_clickup_id,
            username,
            time_estimate,
            time_spent
        FROM tasks
        WHERE 1=1
    `

    args := []interface{}{}
    i := 1


    if startMs != nil {
        query += fmt.Sprintf(`
            AND start_date IS NOT NULL
            AND start_date >= $%d
        `, i)
        args = append(args, *startMs)
        i++
    }

    if endMs != nil {
        query += fmt.Sprintf(`
            AND start_date IS NOT NULL
            AND start_date <= $%d
        `, i)
        args = append(args, *endMs)
        i++
    }
    if username != "" {
        query += fmt.Sprintf(" AND assignee_username ILIKE $%d", i)
        args = append(args, "%"+username+"%")
        i++
    }

    if status != "" {
        query += fmt.Sprintf(" AND status_name = $%d", i)
        args = append(args, status)
        i++
    }

    query += " ORDER BY start_date ASC"

    log.Println("==== GetTasksByRange Query ====")
    log.Println(query)
    log.Printf("ARGS: %+v\n", args)
    log.Println("================================")

    rows, err := r.DB.QueryContext(ctx, query, args...)
    if err != nil {
        log.Println("Query ERROR:", err)
        return nil, err
    }
    defer rows.Close()

    var list []model.Task

    for rows.Next() {
        var t model.Task

        err := rows.Scan(
            &t.ID,
            &t.TaskID,
            &t.Name,
            &t.TextContent,
            &t.Description,
            &t.Status.ID,
            &t.Status.Name,
            &t.Status.Type,
            &t.Status.Color,
            &t.DateDone,
            &t.DateClosed,
            &t.StartDate,
            &t.DueDate,
            &t.AssigneeUsername,
            &t.AssigneeEmail,
            &t.AssigneeColor,
            &t.AssigneeUserID,
            &t.AssigneeClickUpID,
            &t.Username,
            &t.TimeEstimate,
            &t.TimeSpentMs,
        )

        if err != nil {
            log.Println("Scan ERROR:", err)
            return nil, err
        }

        list = append(list, t)
    }

    log.Printf("=== FOUND %d TASKS ===", len(list))

    for _, tt := range list {
        log.Printf(
            "TASK: %-40s | UID: %v | UN: %-20s | EMAIL: %-30s | START: %v | DUE: %v | SPENT: %v",
            tt.Name,
            tt.AssigneeUserID,
            tt.AssigneeUsername,
            tt.AssigneeEmail,
            tt.StartDate,
            tt.DueDate,
            tt.TimeSpentMs,
        )
    }

    return list, nil
}




func (r *PostgresRepo) GetTasksFull(
    ctx context.Context,
    startMs *int64,
    endMs *int64,
    role string,
    username string,
    status string,
) ([]model.TaskFull, error) {

    q := `
        SELECT
            t.id,
            t.name,
            t.description,
            t.status_name,
            t.status_type,
            t.status_color,

            t.start_date,
            t.due_date,
            t.date_done,
            t.date_closed,
            t.time_estimate,
            t.time_spent_ms
        FROM tasks t
        WHERE 1=1
    `

    args := []interface{}{}
    i := 1

    if startMs != nil && endMs != nil {
        q += fmt.Sprintf(" AND t.start_date BETWEEN $%d AND $%d", i, i+1)
        args = append(args, *startMs, *endMs)
        i += 2
    }

    if role != "" {
        q += fmt.Sprintf(" AND m.role = $%d", i)
        args = append(args, role)
        i++
    }

    if username != "" {
        q += fmt.Sprintf(" AND (t.assignee_username ILIKE $%d)", i)
        args = append(args, "%"+username+"%")
        i += 2
    }

    if status != "" {
        q += fmt.Sprintf(" AND t.status_name = $%d", i)
        args = append(args, status)
        i++
    }

    q += " ORDER BY COALESCE(t.start_date, t.date_done, 0) DESC"

    rows, err := r.DB.QueryContext(ctx, q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.TaskFull

    for rows.Next() {

        var tf model.TaskFull

        var (
            startDate, dueDate, dateDone, dateClosed, timeEstimate, timeSpent sql.NullInt64
        )

        err := rows.Scan(
            &tf.TaskID,
            &tf.TaskName,
            &tf.Description,
            &tf.StatusName,
            &tf.StatusType,
            &tf.StatusColor,

            &startDate,
            &dueDate,
            &dateDone,
            &dateClosed,
            &timeEstimate,
            &timeSpent,
        )
        if err != nil {
            return nil, err
        }

        if startDate.Valid { v := startDate.Int64; tf.StartDate = &v }
        if dueDate.Valid { v := dueDate.Int64; tf.DueDate = &v }
        if dateDone.Valid { v := dateDone.Int64; tf.DateDone = &v }
        if dateClosed.Valid { v := dateClosed.Int64; tf.DateClosed = &v }
        if timeEstimate.Valid { v := timeEstimate.Int64; tf.TimeEstimate = &v }
        if timeSpent.Valid { v := timeSpent.Int64; tf.TimeSpent = &v }

        out = append(out, tf)
    }

    return out, nil
}

func (r *PostgresRepo) GetTasksFiltered(ctx context.Context, startDate, endDate *time.Time) ([]*model.TaskResponse, error) {
    query := `
        SELECT 
            id,
            task_id,
            name,
            text_content,
            description,
            status_id,
            status_name,
            status_type,
            status_color,
            date_done,
            date_closed,
            assignee_user_id,
            assignee_clickup_id,
            assignee_username,
            assignee_email,
            start_date,
            due_date,
            time_estimate,
            created_at,
            updated_at
        FROM tasks
        WHERE 
        (
            (start_date IS NOT NULL AND start_date >= $1 AND start_date <= $2)
            OR
            (date_done IS NOT NULL AND date_done >= $1 AND date_done <= $2)
        )
        ORDER BY start_date DESC NULLS LAST, date_done DESC NULLS LAST
    `

    rows, err := r.DB.QueryContext(ctx, query, startDate, endDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    tasks := []*model.TaskResponse{}

    for rows.Next() {
        t := model.TaskResponse{}
        err := rows.Scan(
        &t.ID,
        &t.TaskID,
        &t.Name,
        &t.TextContent,
        &t.Description,
        &t.Status.ID,
        &t.Status.Name,
        &t.Status.Type,
        &t.Status.Color,
        &t.DateDone,
        &t.DateClosed,
        &t.AssigneeUserID,
        &t.AssigneeClickUpID,
        &t.AssigneeUsername,
        &t.AssigneeEmail,
        &t.StartDate,
        &t.DueDate,
        &t.TimeEstimate,
        &t.DateCreated,
        &t.Category,
        &t.Username,
        &t.Email,
        &t.Color,
        &t.TimeEstimateMs,
        &t.TimeSpentMs,
        )

        if err != nil {
            return nil, err
        }

        tasks = append(tasks, &t)
    }

    return tasks, nil
}


func (r *PostgresRepo) GetWorkload(ctx context.Context, start, end time.Time) ([]model.WorkloadUser, error) {
    query := `
        SELECT 
            u.clickup_id AS user_id,
            u.username,
            u.email,
            u.role,
            '' as color,
            COALESCE(SUM(t.time_spent), 0) AS total_ms,
            COUNT(t.id) AS task_count,
            (
                SELECT COUNT(*) 
                FROM tasks tt 
                WHERE tt.assignee_user_id = u.clickup_id
                  AND tt.date_closed BETWEEN $1 AND $2
            ) AS total_tasks
        FROM users u
        LEFT JOIN tasks t 
            ON t.assignee_user_id = u.clickup_id
           AND t.date_closed BETWEEN $1 AND $2
        GROUP BY u.clickup_id, u.username, u.email, u.role
        ORDER BY u.username ASC;
    `

    rows, err := r.DB.QueryContext(ctx, query, start, end)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.WorkloadUser
    for rows.Next() {
        var u model.WorkloadUser
        if err := rows.Scan(
            &u.UserID,
            &u.Username,
            &u.Email,
            &u.Role,
            &u.Color, // Tetap scan, karena query memberikan nilai kosong
            &u.TotalMs,
            &u.TaskCount,
            &u.TotalTasks,
        ); err != nil {
            return nil, err
        }

        u.TotalHours = float64(u.TotalMs) / 3600000.0
        out = append(out, u)
    }

    return out, nil
}

func (r *PostgresRepo) GetTasksByUser(ctx context.Context, userID int64, start, end time.Time) ([]model.TaskItem, error) {
    query := `
        SELECT 
            t.id,
            t.task_id,
            t.name,
            t.description,
            t.text_content,
            t.status_id,
            s.name AS status_name,
            s.type AS status_type,
            s.color AS status_color,
            t.date_done,
            t.date_closed,
            t.time_spent,
            c.name AS category
        FROM tasks t
        LEFT JOIN statuses s ON s.id = t.status_id
        LEFT JOIN categories c ON c.id = t.category_id
        WHERE t.assignee_user_id = $1
          AND t.date_closed BETWEEN $2 AND $3
        ORDER BY t.date_closed DESC;
    `

    rows, err := r.DB.QueryContext(ctx, query, userID, start, end)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []model.TaskItem
    for rows.Next() {
        var t model.TaskItem
        if err := rows.Scan(
            &t.ID,
            &t.TaskID,
            &t.Name,
            &t.Description,
            &t.TextContent,
            &t.StatusID,
            &t.StatusName,
            &t.StatusType,
            &t.StatusColor,
            &t.DateDone,
            &t.DateClosed,
            &t.TimeSpent,
            &t.Category,
        ); err != nil {
            return nil, err
        }

        tasks = append(tasks, t)
    }

    return tasks, nil
}