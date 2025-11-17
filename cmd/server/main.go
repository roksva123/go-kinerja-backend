package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/roksva123/go-kinerja-backend/internal/api/handlers"
	"github.com/roksva123/go-kinerja-backend/internal/config"
	"github.com/roksva123/go-kinerja-backend/internal/middleware"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
	"github.com/roksva123/go-kinerja-backend/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg := config.LoadFromEnv()

	fmt.Println("Connecting to DB...", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName)
	dbCfg := &repository.DBConfig{
		Host: cfg.DBHost,
		Port: cfg.DBPort,
		User: cfg.DBUser,
		Pass: cfg.DBPass,
		Name: cfg.DBName,
	}

	repo, err := repository.NewPostgresRepo(dbCfg)
	if err != nil {
		log.Fatal("DB connect error:", err)
	}

	clickService := service.NewClickUpService(repo, cfg.ClickUpToken, cfg.ClickUpTeamID)

	authHandler := handlers.NewAuthHandler(repo, cfg.JWTSecret)
	employeeHandler := handlers.NewEmployeeHandler(repo, clickService)

	app := gin.Default()

	v1 := app.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.Login)

		protected := v1.Group("")
		protected.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/employees", employeeHandler.ListEmployees)
			protected.GET("/employees/:id", employeeHandler.GetEmployee)
			protected.GET("/employees/:id/tasks", employeeHandler.GetEmployeeTasks)
			protected.GET("/employees/:id/performance", employeeHandler.GetEmployeePerformance)
			protected.GET("/employees/:id/schedule", employeeHandler.GetEmployeeSchedule)
			protected.POST("/sync/clickup", employeeHandler.SyncClickUp)
		}
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	app.Run(":" + port)
}
