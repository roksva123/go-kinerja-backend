package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type AuthHandler struct {
	Repo      *repository.PostgresRepo
	JWTSecret string
}

func NewAuthHandler(repo *repository.PostgresRepo, secret string) *AuthHandler {
	return &AuthHandler{Repo: repo, JWTSecret: secret}
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	admin, err := h.Repo.GetAdminByUsername(context.Background(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": admin.ID,
		"exp": time.Now().Add(10 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	signed, _ := token.SignedString([]byte(h.JWTSecret))
	c.JSON(http.StatusOK, gin.H{"token": signed})
}
