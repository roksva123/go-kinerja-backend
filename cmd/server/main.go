package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/roksva123/go-kinerja-backend/internal/api/handlers"
	"github.com/roksva123/go-kinerja-backend/internal/config"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
	"github.com/roksva123/go-kinerja-backend/internal/service"
	"golang.org/x/crypto/bcrypt"
)

func main() {

	// LOAD ENV
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed load config:", err)
	}

	// DEBUG TOKEN
	fmt.Println("==================================")
	fmt.Println("CLICKUP TOKEN RAW:", cfg.ClickUpToken)
	fmt.Println("TOKEN LENGTH:", len(cfg.ClickUpToken))
	fmt.Println("==================================")


	// INIT DB
	repo := repository.NewPostgresRepo()


	// MIGRATIONS
	if err := repo.RunMigrations(context.Background()); err != nil {
		log.Fatal("migration error:", err)
	}

	// ADMIN SEED
	hashed, _ := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err := repo.UpsertAdmin(context.Background(), cfg.AdminUsername, string(hashed)); err != nil {
		log.Println("failed seeding admin:", err)
	} else {
		log.Println("admin seeded OK")
	}

	// SERVICES
	clickService := service.NewClickUpService(repo, cfg.ClickUpToken, cfg.ClickUpTeamID)

	// HANDLERS
	authHandler := handlers.NewAuthHandler(repo, cfg.JWTSecret)
	employeeHandler := handlers.NewEmployeeHandler(repo, clickService, cfg)
	clickupHandler := handlers.NewClickUpHandler(clickService)

	// ROUTER
	r := gin.Default()
    r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"*"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
    }))
	r.Static("/images", "public/images")
	api := r.Group("/api/v1")

	// CLICKUP ROUTES
	clickup := api.Group("/clickup")
    {
        clickup.POST("/sync/team", clickupHandler.SyncTeam)
        clickup.POST("/sync/members", clickupHandler.SyncMembers)
        clickup.POST("/sync/tasks", clickupHandler.SyncTasks)
    
        clickup.GET("/teams", clickupHandler.GetTeams)
        clickup.GET("/members", clickupHandler.GetMembers)
        clickup.GET("/tasks", clickupHandler.GetTasks)
    }

	// AUTH ROUTES
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// EMPLOYEE + SYNC CLICKUP ROUTES
	api.POST("/sync/clickup", employeeHandler.SyncClickUp)

	emp := api.Group("/employees")
	{
		emp.GET("", employeeHandler.ListEmployees)
		emp.GET("/:id", employeeHandler.GetEmployee)
		emp.GET("/:id/tasks", employeeHandler.GetEmployeeTasks)
		emp.GET("/:id/performance", employeeHandler.GetEmployeePerformance)
		emp.GET("/:id/schedule", employeeHandler.GetEmployeeSchedule)
	}

	// START SERVER
	log.Println("Server running on port:", cfg.Port)
	r.Run(":" + cfg.Port)
}
