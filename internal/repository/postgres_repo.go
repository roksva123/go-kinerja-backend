package repository

import (
    "context"
    "database/sql"
    "time"
    "fmt"
    "os"

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
        `CREATE TABLE IF NOT EXISTS teams (
            team_id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            parent_id TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
        );`,
        `CREATE TABLE IF NOT EXISTS users (
            id BIGINT PRIMARY KEY,
            username TEXT,
            name TEXT,
            password TEXT,
            email TEXT,
            role TEXT,
            color TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
        );`,
        `CREATE TABLE IF NOT EXISTS members (
            id BIGSERIAL PRIMARY KEY,
            clickup_id TEXT UNIQUE,
            username TEXT,
            name TEXT,
            email TEXT,
            color TEXT,
            team_id TEXT REFERENCES teams(team_id) ON DELETE SET NULL,
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
            assignee_clickup_id TEXT,
            assignee_user_id BIGINT,
            assignee_username TEXT,
            assignee_email TEXT,
            assignee_color TEXT,
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

func (r *PostgresRepo) UpsertAdmin(ctx context.Context, username, passwordHash string) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO admins (username, password_hash) VALUES ($1,$2)
		ON CONFLICT (username) DO UPDATE SET password_hash = $2
	`, username, passwordHash)
	return err
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

// UpsertUser (clickup user as local users)
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
    `,
        u.ClickUpID,
        u.Username,
        u.Name,
        u.Email,
        u.PasswordHash,
        u.Role,
        u.Color,
    )
    return err
}


// UpsertMember (if you want members separate)
func (r *PostgresRepo) UpsertMember(ctx context.Context, clickupID, username, name, email, color, teamID string) error {
    _, err := r.DB.ExecContext(ctx, `
        INSERT INTO members (clickup_id, username, name, email, color, team_id)
        VALUES ($1,$2,$3,$4,$5,$6)
        ON CONFLICT (clickup_id) DO UPDATE SET username=EXCLUDED.username, name=EXCLUDED.name, email=EXCLUDED.email, color=EXCLUDED.color, team_id=EXCLUDED.team_id, updated_at=now()
    `, clickupID, username, name, email, color, teamID)
    return err
}

// UpsertTask
func (r *PostgresRepo) UpsertTask(ctx context.Context, t *model.TaskResponse) error {
    var dateDone interface{}
    var dateClosed interface{}
    if t.DateDone != nil {
        dateDone = *t.DateDone
    } else {
        dateDone = nil
    }
    if t.DateClosed != nil {
        dateClosed = *t.DateClosed
    } else {
        dateClosed = nil
    }

    _, err := r.DB.ExecContext(ctx, `
        INSERT INTO tasks (
            id, name, text_content, description,
            status_id, status_name, status_type, status_color,
            date_done, date_closed,
            assignee_clickup_id, assignee_username, assignee_email, assignee_color
        ) VALUES (
            $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14
        ) ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            text_content = EXCLUDED.text_content,
            description = EXCLUDED.description,
            status_id = EXCLUDED.status_id,
            status_name = EXCLUDED.status_name,
            status_type = EXCLUDED.status_type,
            status_color = EXCLUDED.status_color,
            date_done = EXCLUDED.date_done,
            date_closed = EXCLUDED.date_closed,
            assignee_clickup_id = EXCLUDED.assignee_clickup_id,
            assignee_username = EXCLUDED.assignee_username,
            assignee_email = EXCLUDED.assignee_email,
            assignee_color = EXCLUDED.assignee_color,
            updated_at = now()
    `, t.ID, t.Name, t.TextContent, t.Description,
        t.Status.ID, t.Status.Name, t.Status.Type, t.Status.Color,
        dateDone, dateClosed,
        t.Username, t.Username, t.Email, t.Color)
    return err
}

// GetTasks -> returns TaskResponse slice
func (r *PostgresRepo) GetTasks(ctx context.Context) ([]model.TaskResponse, error) {
    q := `
    SELECT id,name,text_content,description,
           status_id,status_name,status_type,status_color,
           date_done,date_closed,
           assignee_username,assignee_email,assignee_color
    FROM tasks
    ORDER BY COALESCE(date_done,0) DESC NULLS LAST
    `
    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    out := []model.TaskResponse{}
    for rows.Next() {
        var t model.TaskResponse
        var statusID, statusName, statusType, statusColor sql.NullString
        var dateDone, dateClosed sql.NullInt64
        var uname, uemail, ucolor sql.NullString

        if err := rows.Scan(
            &t.ID,
            &t.Name,
            &t.TextContent,
            &t.Description,
            &statusID, &statusName, &statusType, &statusColor,
            &dateDone, &dateClosed,
            &uname, &uemail, &ucolor,
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

// GetMembers
func (r *PostgresRepo) GetMembers(ctx context.Context) ([]model.User, error) {
    q := `SELECT id, username, name, email, role, color, created_at, updated_at FROM users ORDER BY name`
    rows, err := r.DB.QueryContext(ctx, q)
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
