package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
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
		`CREATE TABLE IF NOT EXISTS admins (
         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
         username VARCHAR(100) UNIQUE NOT NULL,
         password_hash TEXT NOT NULL,
         created_at TIMESTAMP DEFAULT NOW()
       );`,
        `CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,
        `CREATE TABLE IF NOT EXISTS roles (
            id SERIAL PRIMARY KEY,
            name TEXT UNIQUE NOT NULL
        );`,
        `CREATE TABLE IF NOT EXISTS user_statuses (
            id SERIAL PRIMARY KEY,
            name TEXT UNIQUE NOT NULL
        );`,
        `INSERT INTO roles (name) VALUES
            ('infra'),
            ('mobile apps'),
            ('web'),
            ('backend'),
            ('pm'),
            ('backend-web'),
            ('analis'),
            ('UI-UX')
        ON CONFLICT (name) DO NOTHING;`,
        `INSERT INTO user_statuses (name) VALUES
            ('aktif'),
            ('nonaktif')
        ON CONFLICT (name) DO NOTHING;`,
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
         email TEXT,
         password TEXT,
         role_id INT REFERENCES roles(id) ON DELETE SET NULL,
         status_id INT REFERENCES user_statuses(id) ON DELETE SET NULL,
         created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
         updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
        );`,
        `CREATE TABLE IF NOT EXISTS task_statuses (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            type TEXT,
            color TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
            UNIQUE(id)
        );`,
        `CREATE TABLE IF NOT EXISTS tasks (
        id TEXT PRIMARY KEY,
        name TEXT,
        text_content TEXT,
        description TEXT,
        status_id TEXT REFERENCES task_statuses(id) ON DELETE SET NULL,
        date_done TIMESTAMPTZ,
        date_closed TIMESTAMPTZ,
        start_date TIMESTAMPTZ,
        due_date TIMESTAMPTZ,
        time_estimate_hours FLOAT,
        time_spent_hours FLOAT,
        list_id TEXT,
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
    `DO $$ BEGIN
        ALTER TABLE tasks ADD COLUMN IF NOT EXISTS list_id TEXT;
    END $$;`,
    `DO $$ BEGIN
        ALTER TABLE tasks ADD COLUMN IF NOT EXISTS time_estimate_hours FLOAT;
        ALTER TABLE tasks ADD COLUMN IF NOT EXISTS time_spent_hours FLOAT;
    END $$;`,
    `CREATE INDEX IF NOT EXISTS idx_tasks_start_date ON tasks(start_date);
     CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
     CREATE INDEX IF NOT EXISTS idx_tasks_status_id ON tasks(status_id);`,
    `CREATE TABLE IF NOT EXISTS folders (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        archived BOOLEAN DEFAULT FALSE,
        space_id TEXT,
        created_at TIMESTAMPTZ DEFAULT now(),
        updated_at TIMESTAMPTZ DEFAULT now()
    );`,
    `CREATE TABLE IF NOT EXISTS spaces (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        archived BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMPTZ DEFAULT now(),
        updated_at TIMESTAMPTZ DEFAULT now()
    );`,
    `CREATE TABLE IF NOT EXISTS lists (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        archived BOOLEAN DEFAULT FALSE,
        folder_id TEXT,
        space_id TEXT,
        created_at TIMESTAMPTZ DEFAULT now(),
        updated_at TIMESTAMPTZ DEFAULT now()
    );`,
    `CREATE TABLE IF NOT EXISTS sync_history (
        id SERIAL PRIMARY KEY,
        sync_time TIMESTAMPTZ DEFAULT now(),
        sync_type TEXT NOT NULL,
        status TEXT NOT NULL,
        duration_ms BIGINT,
        details JSONB
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
func (r *PostgresRepo) GetAllUsers(ctx context.Context) ([]model.User, error) { 
    query := `
		SELECT 
			u.clickup_id,
			u.name,
			u.email,
			COALESCE(r.name, '') as role, 
			COALESCE(us.name, '') as status,
			'' as passwordhash,
			u.created_at, 
			u.updated_at 
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		LEFT JOIN user_statuses us ON u.status_id = us.id
		ORDER BY u.name
	`

    rows, err := r.DB.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var users []model.User
    for rows.Next() {
        var u model.User
        if err := rows.Scan(
            &u.ClickUpID, &u.Name, &u.Email, &u.Role, &u.Status, &u.PasswordHash,
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
	defer tx.Rollback() 

	
	_, err = tx.ExecContext(ctx, "DELETE FROM task_assignees WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete old assignees: %w", err)
	}

	if len(assigneeIDs) == 0 {
		return tx.Commit() 
	}

	
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
	log.Printf("UpsertTeam is deprecated, use UpsertSpace instead. Called with teamID: %s", teamID)
	return nil
}

// UpsertUser
func (r *PostgresRepo) UpsertUser(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (clickup_id, name, email, role_id, status_id)
		VALUES (
			$1, $2, $3,
			(SELECT id FROM roles WHERE name ILIKE $4 LIMIT 1),
			(SELECT id FROM user_statuses WHERE name ILIKE $5 LIMIT 1)
		)
		ON CONFLICT (clickup_id) DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			role_id = COALESCE((SELECT id FROM roles WHERE name ILIKE $6 LIMIT 1), users.role_id),
			status_id = (SELECT id FROM user_statuses WHERE name ILIKE $7 LIMIT 1),
			updated_at = now()
	`
	_, err := r.DB.ExecContext(ctx, query, u.ClickUpID, u.Name, u.Email, u.Role, u.Status, u.Role, u.Status)
	if err != nil {
		log.Printf("Error upserting user: %v", err)
	}
	return err
}

func (r *PostgresRepo) UpsertSpace(ctx context.Context, space *model.SpaceInfo) error {
	query := `
		INSERT INTO spaces (id, name)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = now();
	`
	_, err := r.DB.ExecContext(ctx, query, space.ID, space.Name)
	return err
}

func (r *PostgresRepo) UpsertFolder(ctx context.Context, folder *model.Folder) error {
	query := `
		INSERT INTO folders (id, name, archived, space_id)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			archived = EXCLUDED.archived,
			space_id = COALESCE(NULLIF(EXCLUDED.space_id, ''), folders.space_id),
			updated_at = now();
	`
	_, err := r.DB.ExecContext(ctx, query, folder.ID, folder.Name, folder.Archived, folder.Space.ID)
	return err
}

func (r *PostgresRepo) UpsertList(ctx context.Context, list *model.List) error {
	query := `
		INSERT INTO lists (id, name, folder_id, space_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			folder_id = EXCLUDED.folder_id,
			space_id = EXCLUDED.space_id,
			updated_at = now();
	`
	_, err := r.DB.ExecContext(ctx, query, list.ID, list.Name, list.FolderID, list.SpaceID)
	return err
}

func (r *PostgresRepo) UpsertUserFromTask(ctx context.Context, u *model.User) error {
	u.Status = "aktif" 
	return r.UpsertUser(ctx, u)
}

func (r *PostgresRepo) UpsertTaskStatus(ctx context.Context, status *model.TaskStatus) error {
	query := `
		INSERT INTO task_statuses (id, name, type, color)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			color = EXCLUDED.color,
			updated_at = now();
	`
	_, err := r.DB.ExecContext(ctx, query, status.ID, status.Name, status.Type, status.Color)
	return err
}

// UpsertTask
func (r *PostgresRepo) UpsertTask(ctx context.Context, t *model.TaskResponse) error { 
    query := `
        INSERT INTO tasks (
            id, name, text_content, description,
            status_id, date_done, date_closed, start_date, due_date, 
            time_estimate_hours, time_spent_hours, list_id
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (id)
        DO UPDATE SET
            name = EXCLUDED.name,
            text_content = EXCLUDED.text_content,
            description = EXCLUDED.description,
            status_id = EXCLUDED.status_id,
            date_done = EXCLUDED.date_done,
            date_closed = EXCLUDED.date_closed,
            start_date = EXCLUDED.start_date,
            due_date = EXCLUDED.due_date,
            time_estimate_hours = EXCLUDED.time_estimate_hours,
            time_spent_hours = EXCLUDED.time_spent_hours,
            list_id = EXCLUDED.list_id,
            updated_at = now()
    `
    _, err := r.DB.ExecContext(ctx, query,
        t.ID,
        t.Name,
        t.TextContent,
        t.Description,
        t.Status.ID, 
        t.DateDone,      
        t.DateClosed,    
        t.StartDate,     
        t.DueDate,       
        t.TimeEstimateHours,
        t.TimeSpentHours,
        t.ListID,
    )

    return err
}

func (r *PostgresRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
    query := `
        SELECT u.clickup_id, u.name, u.email, COALESCE(r.name, '') as role
        FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
        WHERE u.email = $1
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
            description,
            ts.id as status_id,
            ts.name as status_name,
            ts.type as status_type,
            date_done,    -- Sekarang TIMESTAMPTZ
            start_date,   -- Sekarang TIMESTAMPTZ
            due_date,     -- Sekarang TIMESTAMPTZ
            date_closed,  -- Sekarang TIMESTAMPTZ
            '' as assignee_username, '' as assignee_email, '' as assignee_color
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
            statusID, statusName, statusType sql.NullString
            dateDone, dateClosed, startDate, dueDate       sql.NullTime 
            uname, uemail, ucolor sql.NullString
        )

        if err := rows.Scan(
            &t.ID,
            &t.Name,
            &t.Description,
            &statusID,
            &statusName,
            &statusType,
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

        if dateDone.Valid { t.DateDone = &dateDone.Time }
        if dateClosed.Valid { t.DateClosed = &dateClosed.Time }
        if dueDate.Valid { t.DueDate = &dueDate.Time }
        if startDate.Valid { t.StartDate = &startDate.Time }
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
    q := `
		SELECT 
			u.clickup_id,
			u.name,
			u.email, 
			COALESCE(r.name, '') as role, 
			COALESCE(us.name, '') as status, 
			'' as passwordhash,
			u.created_at, 
			u.updated_at 
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		LEFT JOIN user_statuses us ON u.status_id = us.id
		ORDER BY u.name
	`
    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []model.User
    for rows.Next() {
        var u model.User
        if err := rows.Scan(
			&u.ClickUpID, 
			&u.Name,
			&u.Email,
			&u.Role,
			&u.Status,
			&u.PasswordHash,
			&u.CreatedAt,
			&u.UpdatedAt); err != nil {
            return nil, err
        }
        out = append(out, u)
    }
    return out, nil
}

// GetSpaces
func (r *PostgresRepo) GetSpaces(ctx context.Context) ([]model.SpaceInfo, error) {
    q := `SELECT id, name FROM spaces ORDER BY name`
    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []model.SpaceInfo
    for rows.Next() {
        var s model.SpaceInfo
        if err := rows.Scan(&s.ID, &s.Name); err != nil {
            return nil, err
        }
        out = append(out, s)
    }
    return out, nil
}



// GetFullSyncFiltered 
func (r *PostgresRepo) GetFullSyncFiltered(ctx context.Context, start, end *int64, role string) ([]model.TaskWithMember, error) {

    q := `
        SELECT 
            t.id, t.name, ts.name as status_name,
            t.start_date, t.date_done, t.date_closed,
            u.clickup_id, u.name, u.email, COALESCE(r.name, '')
        FROM tasks t
        LEFT JOIN task_assignees ta ON t.id = ta.task_id
        LEFT JOIN users u ON ta.user_clickup_id = u.clickup_id
        LEFT JOIN task_statuses ts ON t.status_id = ts.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE 1=1
    ` 

    args := []interface{}{}
    idx := 1

    if start != nil && end != nil {
        q += fmt.Sprintf(`
            AND (
                (t.start_date IS NOT NULL AND t.start_date BETWEEN to_timestamp($%[1]d / 1000.0) AND to_timestamp($%[2]d / 1000.0)) OR
                (t.date_done IS NOT NULL AND t.date_done BETWEEN to_timestamp($%[1]d / 1000.0) AND to_timestamp($%[2]d / 1000.0))
            )
        `, idx, idx+1, idx, idx+1, idx, idx+1)

        args = append(args, *start, *end)
        idx += 2
    }

    if role != "" {
        q += fmt.Sprintf(" AND r.name = $%d", idx)
        args = append(args, role)
        idx++
    }

    q += " ORDER BY t.start_date DESC "

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
        
        var startDate, dateDone, dateClosed sql.NullTime

        err := rows.Scan(
            &t.TaskID, &t.TaskName, &t.TaskStatus,
            &startDate, &dateDone, &dateClosed,
            &mID, &mU, &mE, &mC,
        )
        if err != nil {
            return nil, err
        }

        if mID.Valid {
            if startDate.Valid { t.StartDate = &startDate.Time }
            if dateDone.Valid { t.DateDone = &dateDone.Time }
            if dateClosed.Valid { t.DateClosed = &dateClosed.Time }
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
            ts.name as status_name, ts.type as status_type,
            t.start_date, t.due_date, t.date_done, t.date_closed, t.time_estimate_hours, t.time_spent_hours,
            u.clickup_id, u.username, u.email, COALESCE(r.name, '') as role
        FROM tasks t
        LEFT JOIN task_assignees ta ON t.id = ta.task_id
        LEFT JOIN users u ON ta.user_clickup_id = u.clickup_id
        LEFT JOIN task_statuses ts ON t.status_id = ts.id
        LEFT JOIN roles r ON u.role_id = r.id
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
        q += fmt.Sprintf(" AND (u.name ILIKE $%d)", idx)
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
        var startDate, dueDate, dateDone, dateClosed sql.NullTime
        var timeEstimate, timeSpent sql.NullFloat64

        err := rows.Scan(
            &t.TaskID, &t.TaskName, &t.TaskDescription,
            &t.TaskStatus, &t.TaskStatusType,
            &startDate, &dueDate, &dateDone, &dateClosed, &timeEstimate, &timeSpent, // Changed to NullFloat64
            &userID, &userUsername, &userEmail, &userRole,
        )
        if err != nil {
            return nil, err
        }

        if startDate.Valid { t.StartDate = &startDate.Time }
        if dueDate.Valid { t.DueDate = &dueDate.Time }
        if dateDone.Valid { t.DateDone = &dateDone.Time }
        if dateClosed.Valid { t.DateClosed = &dateClosed.Time }
        if timeSpent.Valid { t.TimeSpentHours = &timeSpent.Float64 }
        if timeEstimate.Valid { t.TimeEstimateHours = &timeEstimate.Float64 }

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
    folderIDs []string,
) ([]model.TaskFull, error) {

    query := `
        SELECT
            t.id,
            t.name,
            t.description,
            ts.id as status_id,
            ts.name AS status_name,
            ts.type AS status_type,
            '' AS status_color,
            f.name AS project_name,
            t.date_done,
            t.date_closed,
            t.start_date,
            t.due_date,
            t.time_estimate_hours,
            t.time_spent_hours,
            u.clickup_id,
            u.name AS user_name,
            u.email,
            r.name AS user_role,
            '' AS user_color 
        FROM tasks t
        LEFT JOIN task_assignees ta ON t.id = ta.task_id
        LEFT JOIN users u ON ta.user_clickup_id = u.clickup_id 
        LEFT JOIN lists l ON t.list_id = l.id
        LEFT JOIN folders f ON l.folder_id = f.id
        LEFT JOIN task_statuses ts ON t.status_id = ts.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE 1=1
    `

    args := []interface{}{}
    i := 1

    if startMs != nil && endMs != nil {
        startTime := time.UnixMilli(*startMs)
        endTime := time.UnixMilli(*endMs)
        query += fmt.Sprintf(`
            AND (
                (t.start_date IS NOT NULL AND t.due_date IS NOT NULL AND t.start_date <= $%d AND t.due_date >= $%d) OR
                (t.date_done IS NOT NULL AND t.date_done BETWEEN $%d AND $%d) OR
                (t.date_closed IS NOT NULL AND t.date_closed BETWEEN $%d AND $%d)
            )
        `, i+1, i, i, i+1)
        args = append(args, startTime, endTime)
        i += 2
    }
    if username != "" {
        query += fmt.Sprintf(" AND u.name ILIKE $%d", i)
        args = append(args, "%"+username+"%")
        i++
    }

    if status != "" {
        query += fmt.Sprintf(" AND ts.name = $%d", i)
        args = append(args, status)
        i++
    }

    // Filter by folderIDs (new requirement)
    if len(folderIDs) > 0 {
        folderPlaceholders := make([]string, len(folderIDs))
        for j := range folderIDs {
            folderPlaceholders[j] = fmt.Sprintf("$%d", i+j)
        }
        query += fmt.Sprintf(" AND l.folder_id IN (%s)", strings.Join(folderPlaceholders, ","))
        for _, id := range folderIDs {
            args = append(args, id)
        }
        i += len(folderIDs)
    }

    query += " ORDER BY t.start_date DESC, t.name ASC"

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

    var tasks []model.TaskFull

    for rows.Next() {
        var (
            tf model.TaskFull

            description sql.NullString
            statusID    sql.NullString 
            projectName                               sql.NullString
            startDate, dueDate, dateDone, dateClosed  sql.NullTime
            timeEstimate, timeSpent                   sql.NullFloat64
            userID                                    sql.NullInt64
            assigneeName, assigneeEmail, userRole, userColor sql.NullString
        )

        err := rows.Scan(
            &tf.TaskID,
            &tf.TaskName,
            &description,
            &statusID,      
            &tf.StatusName,
            &tf.StatusType,
            &tf.StatusColor, 
            &projectName,
            &dateDone,
            &dateClosed,
            &startDate,
            &dueDate,
            &timeEstimate,
            &timeSpent,
            &userID,
            &assigneeName,
            &assigneeEmail,
            &userRole,
            &userColor, 
        )

        if err != nil {
            log.Println("Scan ERROR:", err)
            return nil, err
        }
        
        if description.Valid {
            tf.Description = description.String
        }

        if projectName.Valid {
            tf.ProjectName = &projectName.String
        }
        if startDate.Valid {
            tf.StartDate = &startDate.Time
        }
        if dueDate.Valid {
            tf.DueDate = &dueDate.Time
        }
        if dateDone.Valid {
            tf.DateDone = &dateDone.Time
        }
        if dateClosed.Valid {
            tf.DateClosed = &dateClosed.Time
        }
        if userID.Valid {
            tf.UserID = &userID.Int64
        }
        if assigneeName.Valid {
            tf.Username = &assigneeName.String
        }
        if assigneeEmail.Valid {
            tf.Email = &assigneeEmail.String
        }
        if userRole.Valid {
            tf.Role = &userRole.String
        }
        if userColor.Valid {
            tf.Color = &userColor.String
        }

        if timeEstimate.Valid { tf.TimeEstimateHours = timeEstimate.Float64 }
        if timeSpent.Valid { tf.TimeSpentHours = timeSpent.Float64 }

        tasks = append(tasks, tf)
    }

    log.Printf("=== FOUND %d TASKS ===", len(tasks))

    for _, tt := range tasks {
        log.Printf(
            "TASK: %-40s | UID: %v | UN: %-20s | EMAIL: %-30s | START: %v | DUE: %v | SPENT: %v",
            tt.TaskName,
            tt.UserID,
            tt.Username,
            tt.Email,
            tt.StartDate,
            tt.DueDate,
            tt.TimeSpentHours,
        )
    }

    return tasks, nil
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
            ts.name as status_name,
            ts.type as status_type,
            t.start_date,
            t.due_date,
            t.date_done,
            t.date_closed,
            t.time_estimate_hours,
            t.time_spent_hours,
            f.name AS project_name,
            u.clickup_id,
            u.name AS user_name,
            u.email,
            r.name
        FROM tasks t
        LEFT JOIN task_assignees ta ON t.id = ta.task_id
        LEFT JOIN users u ON ta.user_clickup_id = u.clickup_id
        LEFT JOIN lists l ON t.list_id = l.id
        LEFT JOIN folders f ON l.folder_id = f.id
        LEFT JOIN task_statuses ts ON t.status_id = ts.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE 1=1
    `

    args := []interface{}{}
    i := 1

    if startMs != nil && endMs != nil {
        startTime := time.UnixMilli(*startMs)
        endTime := time.UnixMilli(*endMs)
        q += fmt.Sprintf(`
            AND (
                (t.start_date IS NOT NULL AND t.due_date IS NOT NULL AND t.start_date <= $%d AND t.due_date >= $%d) OR
                (t.date_done IS NOT NULL AND t.date_done BETWEEN $%d AND $%d)
            )
        `, i+1, i, i, i+1)
        args = append(args, startTime, endTime)
        i += 2
    }

    if role != "" {
        q += fmt.Sprintf(" AND r.name = $%d", i)
        args = append(args, role)
        i++
    }

    if username != "" {
        q += fmt.Sprintf(" AND u.name ILIKE $%d", i)
        args = append(args, "%"+username+"%")
        i++
    }

    if status != "" {
        q += fmt.Sprintf(" AND ts.name = $%d", i)
        args = append(args, status)
        i++
    }

    q += " ORDER BY t.start_date DESC, t.name ASC"

    rows, err := r.DB.QueryContext(ctx, q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.TaskFull

    for rows.Next() {

        var tf model.TaskFull
		
        var (
            startDate, dueDate, dateDone, dateClosed sql.NullTime
            timeEstimate, timeSpent                  sql.NullFloat64
            userID                                   sql.NullInt64
            assigneeName, userRole, assigneeEmail    sql.NullString
            projectName                              sql.NullString
        )

        err := rows.Scan(
            &tf.TaskID,
            &tf.TaskName,
            &tf.Description,
            &tf.StatusName,
            &tf.StatusType,
            &startDate,
            &dueDate,
            &dateDone,
            &dateClosed,
            &timeEstimate,
            &timeSpent,
            &projectName,
            &userID,
            &assigneeName,
            &assigneeEmail,
            &userRole,
        )
        if err != nil {
            return nil, err
        }

		if startDate.Valid { tf.StartDate = &startDate.Time }
		if dueDate.Valid { tf.DueDate = &dueDate.Time }
		if dateDone.Valid { tf.DateDone = &dateDone.Time }
		if dateClosed.Valid { tf.DateClosed = &dateClosed.Time }
		if userID.Valid {
			tf.UserID = &userID.Int64
		}
		if projectName.Valid {
			tf.ProjectName = &projectName.String
		}
		if assigneeName.Valid {
			tf.Username = &assigneeName.String
		}
		if assigneeEmail.Valid {
			tf.Email = &assigneeEmail.String
		}
		if userRole.Valid {
			tf.Role = &userRole.String
		}

        if timeEstimate.Valid { tf.TimeEstimateHours = timeEstimate.Float64 }
        if timeSpent.Valid { tf.TimeSpentHours = timeSpent.Float64 }

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
            description,
            t.status_id,
            ts.name as status_name,
            ts.type as status_type,
            date_done,
            date_closed,
            assignee_user_id,
            assignee_clickup_id,
            assignee_username,
            assignee_email,
            start_date,
            due_date,
            time_estimate_hours,
            created_at,
            updated_at
        FROM tasks
        LEFT JOIN task_statuses ts ON t.status_id = ts.id
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
        &t.Description,
        &t.Status.ID,
        &t.Status.Name,
        &t.Status.Type,
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
        &t.DateUpdated,
        )

        if err != nil {
            return nil, err
        }

        tasks = append(tasks, &t)
    }

    return tasks, nil
}


func calculateWorkingDays(start, end time.Time) int {
	if start.After(end) {
		return 0
	}

	workingDays := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		weekday := d.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			workingDays++
		}
	}
	return workingDays
}
func (r *PostgresRepo) GetWorkload(ctx context.Context, start, end time.Time) ([]model.WorkloadUser, error) {
    query := `
        SELECT 
            u.clickup_id,
            u.name,
            u.email,
            COALESCE(r.name, '') AS role,
            '' AS color,
            COALESCE(SUM(t.time_spent_hours), 0) AS total_hours,
            COUNT(t.id) FILTER (WHERE t.id IS NOT NULL) AS task_count
        FROM users u
        LEFT JOIN roles r ON u.role_id = r.id
        LEFT JOIN task_assignees ta ON u.clickup_id = ta.user_clickup_id
        LEFT JOIN tasks t ON ta.task_id = t.id AND (
            (t.start_date IS NOT NULL AND t.due_date IS NOT NULL AND t.start_date <= $2 AND t.due_date >= $1) OR
            (t.date_done IS NOT NULL AND t.date_done BETWEEN $1 AND $2) OR
            (t.date_closed IS NOT NULL AND t.date_closed BETWEEN $1 AND $2)
        )
        WHERE u.status_id = (SELECT id FROM user_statuses WHERE name = 'aktif')
        GROUP BY u.clickup_id, u.name, u.email, r.name
        ORDER BY u.name ASC
    `

    rows, err := r.DB.QueryContext(ctx, query, start, end)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

	workingDays := calculateWorkingDays(start, end)
	expectedWorkHours := float64(workingDays * 8)
    var out []model.WorkloadUser
    for rows.Next() {
        var u model.WorkloadUser
        var totalHours sql.NullFloat64
        if err := rows.Scan(
            &u.UserID,
            &u.Name,
            &u.Email,
            &u.Role,
            &u.Color, 
            &totalHours,
            &u.TaskCount,
        ); err != nil {
            return nil, err
        }
        if totalHours.Valid { u.TotalHours = totalHours.Float64 }
		u.ExpectedHours = expectedWorkHours
        out = append(out, u)
    }

    return out, nil
}

func (r *PostgresRepo) GetTasksByUser(ctx context.Context, userID int64, start, end time.Time) ([]model.TaskItem, error) {
    query := `
        SELECT 
            t.id,
            t.name,
            t.description, 
            ts.id as status_id,
            ts.name as status_name,
            ts.type as status_type,
            t.date_done,
            t.date_closed,
            t.start_date,
            t.due_date,
            t.time_estimate_hours,
            t.time_spent_hours,
            COALESCE(f.name, l.name) AS project_name,
            '' AS category
        FROM tasks t
        INNER JOIN task_assignees ta ON t.id = ta.task_id
        LEFT JOIN task_statuses ts ON t.status_id = ts.id
        LEFT JOIN lists l ON t.list_id = l.id
        LEFT JOIN folders f ON l.folder_id = f.id
        WHERE ta.user_clickup_id = $1
          AND (
            (t.start_date IS NOT NULL AND t.due_date IS NOT NULL AND t.start_date <= $3 AND t.due_date >= $2) OR
            (t.date_done IS NOT NULL AND t.date_done BETWEEN $2 AND $3)
        )
        ORDER BY t.start_date DESC, t.name ASC
    `

    rows, err := r.DB.QueryContext(ctx, query, userID, start, end)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []model.TaskItem
    for rows.Next() {
        var t model.TaskItem
        var timeEstimate, timeSpent sql.NullFloat64
        var startDate, dueDate, dateDone, dateClosed sql.NullTime
        var category sql.NullString
        var description sql.NullString
        var projectName sql.NullString

        if err := rows.Scan(
            &t.ID,
            &t.Name,
            &description,
            &t.StatusID,
            &t.StatusName,
            &t.StatusType,
            &dateDone,
            &dateClosed,
            &startDate,
            &dueDate,
            &timeEstimate,
            &timeSpent,
            &projectName,
            &category,
        ); err != nil {
            return nil, err
        }

        if description.Valid { t.Description = description.String }
        if dateDone.Valid { t.DateDone = &dateDone.Time }
        if dateClosed.Valid { t.DateClosed = &dateClosed.Time }
        if startDate.Valid { t.StartDate = &startDate.Time }
        if dueDate.Valid { t.DueDate = &dueDate.Time }
        if timeEstimate.Valid { t.TimeEstimateHours = timeEstimate.Float64 }
        if timeSpent.Valid { t.TimeSpentHours = timeSpent.Float64 }
        if projectName.Valid {
            t.ProjectName = &projectName.String
        }

        tasks = append(tasks, t)
    }

    return tasks, nil
}

func (r *PostgresRepo) GetTasksSummary(
	ctx context.Context,
	startMs *int64,
	endMs *int64,
	username string,
) (*model.TaskSummary, error) {

	query := `
        SELECT
            COUNT(t.id),
            COALESCE(SUM(t.time_spent_hours), 0),
            COALESCE(SUM(CASE WHEN ts.type = 'open' THEN t.time_estimate_hours ELSE 0 END), 0)
        FROM tasks t
        LEFT JOIN task_assignees ta ON t.id = ta.task_id
        LEFT JOIN users u ON ta.user_clickup_id = u.clickup_id
        WHERE 1=1
    `

	args := []interface{}{}
	i := 1

	if startMs != nil && endMs != nil {
		query += fmt.Sprintf(`
            AND (
                (t.start_date <= to_timestamp($%d / 1000.0) AND t.due_date >= to_timestamp($%d / 1000.0)) OR
                (t.date_done >= to_timestamp($%d / 1000.0) AND t.date_done <= to_timestamp($%d / 1000.0)) OR
                (t.date_closed >= to_timestamp($%d / 1000.0) AND t.date_closed <= to_timestamp($%d / 1000.0))
            )
        `, i+1, i, i, i+1, i, i+1) 
		args = append(args, *startMs, *endMs)
		i += 2
	}

	if username != "" {
		query += fmt.Sprintf(" AND u.name ILIKE $%d", i)
		args = append(args, "%"+username+"%")
		i++
	}

	var totalTasks int
	var totalTimeSpent, totalTimeEstimate float64

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(
		&totalTasks,
		&totalTimeSpent,
		&totalTimeEstimate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &model.TaskSummary{}, nil 
		}
		return nil, err
	}

	summary := &model.TaskSummary{
		TotalTasks:         totalTasks,
		ActualWorkHours:    totalTimeSpent,
		TotalUpcomingHours: totalTimeEstimate,
	}

	return summary, nil
}

func (r *PostgresRepo) GetTasksSummaryByDateRange(ctx context.Context, start, end time.Time) ([]model.TaskSummary, error) {
	query := `
		SELECT
			u.clickup_id,
			u.name,
			u.email,
			COUNT(t.id) FILTER (WHERE t.id IS NOT NULL) AS total_tasks,
			COALESCE(SUM(t.time_spent_hours) FILTER (
				WHERE ts.type IN ('closed', 'done', 'complete') OR LOWER(ts.name) LIKE '%done%' OR LOWER(ts.name) LIKE '%completed%'
			), 0) AS actual_work_hours,
			COALESCE(SUM(t.time_estimate_hours) FILTER (
				WHERE ts.type = 'open' OR LOWER(ts.name) LIKE '%to do%'
			), 0) AS total_upcoming_hours
		FROM users u
		LEFT JOIN task_assignees ta ON u.clickup_id = ta.user_clickup_id
		LEFT JOIN tasks t ON ta.task_id = t.id AND (
			(t.start_date IS NOT NULL AND t.due_date IS NOT NULL AND t.start_date <= $2 AND t.due_date >= $1) OR
            (t.date_done IS NOT NULL AND t.date_done BETWEEN $1 AND $2) OR
            (t.date_closed IS NOT NULL AND t.date_closed BETWEEN $1 AND $2)
		)
		LEFT JOIN task_statuses ts ON t.status_id = ts.id
		WHERE u.status_id = (SELECT id FROM user_statuses WHERE name = 'aktif')
		GROUP BY u.clickup_id, u.name, u.email
		ORDER BY u.name ASC;
	`

	rows, err := r.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("querying tasks summary by date range failed: %w", err)
	}
	defer rows.Close()

	var summaries []model.TaskSummary
	workingDays := calculateWorkingDays(start, end)
	expectedWorkHours := float64(workingDays * 8)

	for rows.Next() {
		var s model.TaskSummary
		var totalTimeSpent, totalUpcomingEstimate float64

		if err := rows.Scan(
			&s.UserID,
			&s.Name,
			&s.Email,
			&s.TotalTasks,
			&totalTimeSpent,
			&totalUpcomingEstimate,
		); err != nil {
			return nil, fmt.Errorf("scanning task summary row failed: %w", err)
		}

		s.ActualWorkHours = totalTimeSpent
		s.TotalUpcomingHours = totalUpcomingEstimate
		s.TotalWorkHours = expectedWorkHours

		summaries = append(summaries, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for task summary: %w", err)
	}

	return summaries, nil
}

func (r *PostgresRepo) GetFolders(ctx context.Context) ([]model.Folder, error) {
	query := `
		SELECT f.id, f.name, f.archived, f.space_id, COALESCE(s.name, '') as space_name
		FROM folders f
		LEFT JOIN spaces s ON f.space_id = s.id
		ORDER BY s.name, f.name
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.Archived, &f.Space.ID, &f.Space.Name); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}

	return folders, nil
}

func (r *PostgresRepo) GetLists(ctx context.Context) ([]model.List, error) {
	query := `SELECT id, name, archived, folder_id, space_id FROM lists ORDER BY name`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []model.List
	for rows.Next() {
		var l model.List
		if err := rows.Scan(&l.ID, &l.Name, &l.Archived, &l.FolderID, &l.SpaceID); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}

	return lists, nil
}

func (r *PostgresRepo) CreateSyncHistory(ctx context.Context, syncType, status string, durationMs int64, details []byte) (int64, error) {
	query := `
		INSERT INTO sync_history (sync_type, status, duration_ms, details)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var detailsJSON sql.NullString
	if details != nil {
		detailsJSON.String = string(details)
		detailsJSON.Valid = true
	}

	var historyID int64
	err := r.DB.QueryRowContext(ctx, query, syncType, status, durationMs, detailsJSON).Scan(&historyID)

	return historyID, err
}

func (r *PostgresRepo) GetSyncHistory(ctx context.Context, limit int) ([]model.SyncHistory, error) {
	query := `
		SELECT id, sync_time, sync_type, status, COALESCE(duration_ms, 0), details
		FROM sync_history
		ORDER BY sync_time DESC
		LIMIT $1
	`

	if limit <= 0 {
		limit = 50 
	}

	rows, err := r.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("querying sync history failed: %w", err)
	}
	defer rows.Close()

	var history []model.SyncHistory
	for rows.Next() {
		var h model.SyncHistory
		var details []byte 

		if err := rows.Scan(&h.ID, &h.SyncTime, &h.SyncType, &h.Status, &h.DurationMs, &details); err != nil {
			return nil, fmt.Errorf("scanning sync history row failed: %w", err)
		}

		if details != nil {
			h.Details = details
		}

		history = append(history, h)
	}

	return history, nil
}