package repository

import (
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
        &user.ID, &user.Username, &user.Password,
        &user.Name, &user.Role,
        &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return user, nil
}
