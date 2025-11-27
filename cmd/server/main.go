package main

import (
	"context"
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
	clickSvc := service.NewClickUpService(repo, cfg.ClickUpToken, cfg.ClickUpTeamID)
	workloadSvc := service.NewWorkloadService(repo)
	// authSvc := service.NewAuthService(repo, cfg.JWTSecret) 

	// HANDLERS (constructor diperbaiki)
	clickupHandler := handlers.NewClickUpHandler(clickSvc)
	workloadHandler := handlers.NewWorkloadHandler(workloadSvc)
	authHandler := handlers.NewAuthHandler(repo, cfg.JWTSecret)          

	// ROUTER
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/images", "public/images")
	api := r.Group("/api/v1")

	// CLICKUP ROUTES
	clickup := api.Group("/clickup")
	{
		clickup.POST("/sync/team", clickupHandler.SyncTeam)
		clickup.POST("/sync/members", clickupHandler.SyncMembers)
		clickup.POST("/sync/tasks", clickupHandler.SyncTasks)
		clickup.POST("/fullsync", clickupHandler.FullSync)

		clickup.GET("/teams", clickupHandler.GetTeams)
		clickup.GET("/members", clickupHandler.GetMembers)
		clickup.GET("/tasks", clickupHandler.GetTasks)
		clickup.GET("/fullsync", clickupHandler.GetFullSync)
		clickup.GET("/fullsync/filter", clickupHandler.GetFullSyncFiltered)
		clickup.GET("/data", clickupHandler.GetFullData)
        // clickup.POST("/sync", clickupHandler.SyncAll)
		// clickup.GET("/workload/allsync", clickupHandler.AllSync)
		// clickup.GET("/workload", clickupHandler.GetWorkload)


	}

	// WORKLOAD ROUTES
	work := api.Group("/workload")
	{
		work.POST("/sync", workloadHandler.SyncAll)
		work.GET("", workloadHandler.GetWorkload)
	}

	// AUTH ROUTES
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// START SERVER
	log.Println("Server running on port:", cfg.Port)
	r.Run(":" + cfg.Port)
}
