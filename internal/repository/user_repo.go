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
    query := `SELECT id, username, password, name, role, created_at, updated_at FROM users WHERE username=$1`
    err := r.db.QueryRow(query, username).Scan(
        &user, &user.Name, &user.PasswordHash,
        &user.Name, &user.Role,
        &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return user, nil
}

func (r *userRepo) UpsertUser(ctx context.Context, u *model.User) error {
	query := `
        INSERT INTO users (id, username, name, role)
        VALUES (?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
            username = VALUES(username),
            name = VALUES(name),
            role = VALUES(role)
    `
	_, err := r.db.ExecContext(ctx, query, u.Name, u.Name, u.Role)
	return err
}

