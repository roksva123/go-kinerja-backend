package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type AuthService struct {
    repo  repository.UserRepo
    jwtKey []byte
}

func NewAuthService(repo repository.UserRepo, jwtKey string) *AuthService {
    return &AuthService{repo, []byte(jwtKey)}
}

func (s *AuthService) Login(username, password string) (string, *model.User, error) {
    user, err := s.repo.GetByUsername(username)
    if err != nil {
        return "", nil, errors.New("invalid credentials")
    }

    if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
        return "", nil, errors.New("invalid credentials")
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ClickUpID,
        "role":    user.Role,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenStr, err := token.SignedString(s.jwtKey)
    if err != nil {
        return "", nil, err
    }

    return tokenStr, user, nil
}
