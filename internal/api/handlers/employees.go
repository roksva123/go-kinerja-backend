package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
	"github.com/roksva123/go-kinerja-backend/internal/service"
)

type EmployeeHandler struct {
	Repo           *repository.PostgresRepo
	ClickUpService *service.ClickUpService
}

func NewEmployeeHandler(repo *repository.PostgresRepo, clickUp *service.ClickUpService) *EmployeeHandler {
	return &EmployeeHandler{Repo: repo, ClickUpService: clickUp}
}

func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	list, err := h.Repo.ListEmployees(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, list)
}

func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	emp, err := h.Repo.GetEmployee(context.Background(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "employee not found"})
		return
	}
	c.JSON(200, emp)
}

func (h *EmployeeHandler) GetEmployeeTasks(c *gin.Context) {
	id := c.Param("id")

	var fromPtr, toPtr *time.Time
	from := c.Query("from")
	to := c.Query("to")

	if from != "" {
		t, _ := time.Parse("2006-01-02", from)
		fromPtr = &t
	}
	if to != "" {
		t, _ := time.Parse("2006-01-02", to)
		toPtr = &t
	}

	list, err := h.Repo.ListTasksByEmployee(context.Background(), id, fromPtr, toPtr)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, list)
}

func calculateWorkloadHours(tasks []repository.Task) float64 {
	var total float64
	for _, t := range tasks {
		if t.TimeEstimateSeconds.Valid {
			total += float64(t.TimeEstimateSeconds.Int64) / 3600
		}
	}
	return total
}

func (h *EmployeeHandler) GetEmployeePerformance(c *gin.Context) {
	id := c.Param("id")
	tasks, _ := h.Repo.ListTasksByEmployee(context.Background(), id, nil, nil)

	totalHours := calculateWorkloadHours(tasks)

	var category string
	switch {
	case totalHours <= 35:
		category = "underload"
	case totalHours >= 60:
		category = "overload"
	default:
		category = "normal"
	}

	c.JSON(200, gin.H{
		"total_hours": totalHours,
		"category":    category,
		"tasks":       tasks,
	})
}

func (h *EmployeeHandler) GetEmployeeSchedule(c *gin.Context) {
	id := c.Param("id")
	now := time.Now()

	list, err := h.Repo.ListTasksByEmployee(context.Background(), id, &now, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if len(list) == 0 {
		c.JSON(200, gin.H{"message": "Maaf, belum ada task untuk tanggal yang akan datang."})
		return
	}

	c.JSON(200, list)
}

func (h *EmployeeHandler) SyncClickUp(c *gin.Context) {
	if err := h.ClickUpService.SyncTasks(context.Background()); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Sync ClickUp success"})
}
