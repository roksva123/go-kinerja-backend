package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/roksva123/go-kinerja-backend/internal/api/handlers"
	"github.com/roksva123/go-kinerja-backend/internal/config"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
	"github.com/roksva123/go-kinerja-backend/internal/service"
	"golang.org/x/crypto/bcrypt"
)

func main() {

	// 1. LOAD ENV + CONFIG
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed load config:", err)
	}

	fmt.Println("Server Port      =", cfg.Port)
	fmt.Println("ClickUp Token    =", cfg.ClickUpToken)
	fmt.Println("ClickUp Team ID  =", cfg.ClickUpTeamID)

	// 2. INIT DATABASE
	repo, err := repository.NewPostgresRepo(&repository.DBConfig{
		Host: cfg.DBHost,
		Port: cfg.DBPort,
		User: cfg.DBUser,
		Pass: cfg.DBPass,
		Name: cfg.DBName,
	})
	if err != nil {
		log.Fatal("failed connect db:", err)
	}

	// 3. MIGRATIONS
	if err := repo.RunMigrations(context.Background()); err != nil {
		log.Fatal("migration error:", err)
	}

	// 4. SEED ADMIN
	hashed, _ := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err := repo.UpsertAdmin(context.Background(), cfg.AdminUsername, string(hashed)); err != nil {
		log.Println("failed seeding admin:", err)
	} else {
		log.Println("admin seeded OK")
	}

	// 5. INIT SERVICES
	clickService := service.NewClickUpService(repo, cfg.ClickUpToken, cfg.ClickUpTeamID)

	// 6. INIT HANDLERS
	authHandler := handlers.NewAuthHandler(repo, cfg.JWTSecret)
	employeeHandler := handlers.NewEmployeeHandler(repo, clickService, cfg)

	// 7. ROUTER
	r := gin.Default()

	api := r.Group("/api/v1")

	// ---------- AUTH ----------
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// ---------- SYNC CLICKUP ----------
	api.POST("/sync/clickup", employeeHandler.SyncClickUp)

	// ---------- EMPLOYEES ----------
	emp := api.Group("/employees")
	{
		emp.GET("", employeeHandler.ListEmployees)
		emp.GET("/:id", employeeHandler.GetEmployee)
		emp.GET("/:id/tasks", employeeHandler.GetEmployeeTasks)
		emp.GET("/:id/performance", employeeHandler.GetEmployeePerformance)
		emp.GET("/:id/schedule", employeeHandler.GetEmployeeSchedule)
	}

	// 8. START SERVER
	log.Println("Server running on port:", cfg.Port)
	r.Run(":" + cfg.Port)
}
