package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roksva123/go-kinerja-backend/internal/service"
	// "github.com/roksva123/go-kinerja-backend/internal/model"
)

type WorkloadHandler struct {
	workloadSvc *service.WorkloadService
	clickupSvc  *service.ClickUpService
}

func NewWorkloadHandler(workloadSvc *service.WorkloadService, clickupSvc *service.ClickUpService) *WorkloadHandler {
	return &WorkloadHandler{
		workloadSvc: workloadSvc,
		clickupSvc:  clickupSvc,
	}
}

func (h *WorkloadHandler) GetTasksSummary(c *gin.Context) {
	var startMs, endMs *int64

	var startDate, endDate time.Time
	var err error

	// Parse start_date dari query param
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		t, err := time.Parse("02-01-2006", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use DD-MM-YYYY"})
			return
		}
		startDate = t
		// Set ke awal hari
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		ms := startOfDay.UnixMilli()
		startMs = &ms
	}

	// Parse end_date dari query param
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		t, err := time.Parse("02-01-2006", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use DD-MM-YYYY"})
			return
		}
		endDate = t
		// Set ke akhir hari
		endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		ms := endOfDay.UnixMilli()
		endMs = &ms
	}

	username := c.Query("username")

	dbSummary, err := h.workloadSvc.GetTasksSummary(c.Request.Context(), startMs, endMs, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks summary: " + err.Error()})
		return
	}

	// Hitung ekspektasi jam kerja jika ada rentang tanggal
	var expectedWorkHours float64
	if !startDate.IsZero() && !endDate.IsZero() {
		workingDays := service.WorkingDaysBetween(startDate, endDate)
		expectedWorkHours = float64(workingDays * 8)
	}

	dbSummary.TotalWorkHours = expectedWorkHours

	c.JSON(http.StatusOK, dbSummary)
}

// --- Placeholder untuk fungsi handler lainnya ---

func (h *WorkloadHandler) GetWorkload(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	layout := "02-01-2006"
	start, err := time.Parse(layout, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use DD-MM-YYYY"})
		return
	}
	end, err := time.Parse(layout, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use DD-MM-YYYY"})
		return
	}

	// Convert to Unix milliseconds
	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	users, err := h.clickupSvc.GetWorkload(c.Request.Context(), startMs, endMs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate summary
	totalHoursAll := 0.0
	for _, u := range users {
		totalHoursAll += u.TotalHours
	}
	avgHours := 0.0
	if len(users) > 0 {
		avgHours = totalHoursAll / float64(len(users))
	}

	c.JSON(http.StatusOK, gin.H{
		"start": startStr,
		"end":   endStr,
		"summary": gin.H{
			"total_users": len(users),
			"total_hours": totalHoursAll,
			"avg_hours":   avgHours,
		},
		"users": users,
	})
}

func (h *WorkloadHandler) SyncAll(c *gin.Context) {
	err := h.workloadSvc.SyncAll(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "workload synced"})
}

func (h *WorkloadHandler) AllSync(c *gin.Context) {
	h.SyncAll(c)
}

func (h *WorkloadHandler) GetTasksByRange(c *gin.Context) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	layout := "02-01-2006"
	startDate, err := time.Parse(layout, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use DD-MM-YYYY"})
		return
	}
	endDate, err := time.Parse(layout, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use DD-MM-YYYY"})
		return
	}

	sortOrder := c.DefaultQuery("sort", "desc")

	workingDays := service.WorkingDaysBetween(startDate, endDate)
	expectedHours := float64(workingDays * 8)

	assigneesMap, tasksByAssignee, err := h.workloadSvc.GetTasksByRangeGroupedByAssignee(c.Request.Context(), startDate, endDate, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responseAssignees []AssigneeWithTasks
	for id, assignee := range assigneesMap {
		
		_ = tasksByAssignee[id] 
		totalSpentHours := 0.0  

		responseAssignees = append(responseAssignees, AssigneeWithTasks{
			ClickupID:       int(assignee.ClickUpID),
			Username:        assignee.Username,
			Email:           assignee.Email,
			Name:            assignee.Name,
			Tasks:           []TaskInResponse{}, // Placeholder
			TotalSpentHours: totalSpentHours,
			ExpectedHours:   expectedHours,
		})
	}

	finalResponse := TasksByAssigneeResponse{
		Count:     len(responseAssignees),
		Assignees: responseAssignees,
	}

	c.JSON(http.StatusOK, finalResponse)
}