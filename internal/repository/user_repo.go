package repository

import (
	"context"
	"database/sql"

	"github.com/roksva123/go-kinerja-backend/internal/model"
)

type UserRepo interface {
    GetByUsername(username string) (*model.User, error)
}

type userRepo struct {
    db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepo {
    return &userRepo{db}
}

func (r *userRepo) GetByUsername(username string) (*model.User, error) {
    user := &model.User{}
    // Query to select all relevant columns for model.User
    query := `SELECT clickup_id, name, email, password, role, status, created_at, updated_at FROM users WHERE username=$1`
    err := r.db.QueryRow(query, username).Scan(
        &user.ClickUpID,
        &user.Name,
        &user.Email,
        &user.PasswordHash,
        &user.Role,
        &user.Status,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // User not found
        }
        return nil, err
    }
    return user, nil
}

func (r *userRepo) UpsertUser(ctx context.Context, u *model.User) error {
    // Corrected UpsertUser to match the users table and model.User struct
	query := `
        INSERT INTO users (clickup_id, username, name, email, role, status, password)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (clickup_id) DO UPDATE SET
            username = VALUES(username),
            name = VALUES(name),
            email = EXCLUDED.email,
            role = EXCLUDED.role,
            status = EXCLUDED.status,
            password = EXCLUDED.password, -- Assuming password can be updated on upsert
            updated_at = now()
    `
	_, err := r.db.ExecContext(ctx, query,
		u.ClickUpID,
		u.Name, // Assuming model.User.Name maps to username in DB
		u.Name, // Assuming model.User.Name also maps to name in DB
		u.Email,
		u.Role,
		u.Status,
		u.PasswordHash,
	)
	return err
}
