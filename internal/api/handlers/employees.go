package handlers

import (
	"context"

	"time"

	"github.com/gin-gonic/gin"

	"github.com/roksva123/go-kinerja-backend/internal/config"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
	"github.com/roksva123/go-kinerja-backend/internal/service"
)

type EmployeeHandler struct {
	Repo   *repository.PostgresRepo
	Click  *service.ClickUpService
	Config *config.Config
}

func NewEmployeeHandler(repo *repository.PostgresRepo, click *service.ClickUpService, cfg *config.Config) *EmployeeHandler {
	return &EmployeeHandler{Repo: repo, Click: click, Config: cfg}
}

func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	list, err := h.Repo.ListEmployees(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// map to response (simple)
	out := []map[string]interface{}{}
	for _, e := range list {
		out = append(out, map[string]interface{}{
			"id": e.ID, "fullname": e.Fullname, "email": e.Email.String, "clickup_id": e.ClickUpID.String,
		})
	}
	c.JSON(200, out)
}

func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	e, err := h.Repo.GetEmployee(context.Background(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, gin.H{"id": e.ID, "fullname": e.Fullname, "email": e.Email.String, "clickup_id": e.ClickUpID.String})
}

func (h *EmployeeHandler) GetEmployeeTasks(c *gin.Context) {
	id := c.Param("id")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	var from, to *time.Time
	if fromStr != "" {
		t, _ := time.Parse("2006-01-02", fromStr)
		from = &t
	}
	if toStr != "" {
		t, _ := time.Parse("2006-01-02", toStr)
		to = &t
	}
	tasks, err := h.Repo.ListTasksByEmployee(context.Background(), id, from, to)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, tasks)
}

func (h *EmployeeHandler) GetEmployeePerformance(c *gin.Context) {
	id := c.Param("id")
	// default date range: last 7 days or use query
	fromStr := c.Query("from")
	toStr := c.Query("to")
	var from, to *time.Time
	if fromStr != "" {
		t, _ := time.Parse("2006-01-02", fromStr)
		from = &t
	}
	if toStr != "" {
		t, _ := time.Parse("2006-01-02", toStr)
		to = &t
	}
	tasks, err := h.Repo.ListTasksByEmployee(context.Background(), id, from, to)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// compute total hours (use time_spent if present else time_estimate)
	var totalSeconds int64
	var taskCount int
	var sumPercent float64
	for _, t := range tasks {
		taskCount++
		if t.TimeSpentSeconds.Valid {
			totalSeconds += t.TimeSpentSeconds.Int64
		} else if t.TimeEstimateSeconds.Valid {
			totalSeconds += t.TimeEstimateSeconds.Int64
		}
		if t.PercentComplete.Valid {
			sumPercent += t.PercentComplete.Float64
		}
	}
	totalHours := float64(totalSeconds) / 3600.0
	avgPercent := 0.0
	if taskCount > 0 {
		avgPercent = sumPercent / float64(taskCount)
	}
	category := "normal"
	if totalHours <= h.Config.WorkloadUnderload {
		category = "underload"
	} else if totalHours >= h.Config.WorkloadOverload {
		category = "overload"
	} else {
		// normal band check (36-45)
		if totalHours >= h.Config.WorkloadNormalMin && totalHours <= h.Config.WorkloadNormalMax {
			category = "normal"
		} else {
			// if between underload and normal or between normal and overload
			category = "normal"
		}
	}
	c.JSON(200, gin.H{
		"employee_id": id,
		"total_hours": totalHours,
		"tasks_count": taskCount,
		"avg_percent_complete": avgPercent,
		"category": category,
	})
}

func (h *EmployeeHandler) GetEmployeeSchedule(c *gin.Context) {
	id := c.Param("id")
	dateStr := c.Query("date")
	var date time.Time
	var err error
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid date"})
			return
		}
	}
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	to := from.Add(24*time.Hour - time.Nanosecond)
	tasks, err := h.Repo.ListTasksByEmployee(context.Background(), id, &from, &to)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if len(tasks) == 0 {
		c.JSON(200, gin.H{"message": "Maaf, task belum tersedia"})
		return
	}
	c.JSON(200, tasks)
}

func (h *EmployeeHandler) SyncClickUp(c *gin.Context) {
	ctx := context.Background()
	// sync members then tasks
	if err := h.Click.SyncMembers(ctx); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	n, err := h.Click.SyncTasks(ctx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "sync ok", "tasks_synced": n})
}
