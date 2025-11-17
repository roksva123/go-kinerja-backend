package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type AuthHandler struct {
	Repo      *repository.PostgresRepo
	JWTSecret string
}

func NewAuthHandler(repo *repository.PostgresRepo, jwtSecret string) *AuthHandler {
	return &AuthHandler{Repo: repo, JWTSecret: jwtSecret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	var response model.ResponseApi

	// Validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ApiMessage = "Invalid request: " + err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Fetch admin by username
	admin, err := h.Repo.GetAdminByUsername(context.Background(), req.Username)
	if err != nil {
		response.ApiMessage = "Username or password is incorrect"
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Compare password using bcrypt
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
		response.ApiMessage = "Username or password is incorrect"
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Create JWT token
	claims := jwt.MapClaims{
		"sub": admin.ID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(12 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		response.ApiMessage = "Failed to generate token"
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response.ApiMessage = "Login Successful"
	response.Data = model.LoginResponse{
		Token: tokenString,
	}

	c.JSON(http.StatusOK, response)
}
