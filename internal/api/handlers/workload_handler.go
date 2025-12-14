package handlers

import (
	"fmt"
	"math"
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	name := c.Query("name")
	email := c.Query("email")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	layout := "02-01-2006"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use DD-MM-YYYY"})
		return
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use DD-MM-YYYY"})
		return
	}

	summary, err := h.workloadSvc.GetTasksSummary(c.Request.Context(), startDate, endDate, name, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *WorkloadHandler) GetWorkload(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	username := c.Query("username")


	layout := "2006-01-02" 
	start, err := time.Parse(layout, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use DD-MM-YYYY", "details": err.Error()})
		return
	}
	end, err := time.Parse(layout, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use DD-MM-YYYY", "details": err.Error()})
		return
	}

	users, err := h.workloadSvc.GetWorkload(c.Request.Context(), start, end, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *WorkloadHandler) SyncAll(c *gin.Context) {
	err := h.workloadSvc.SyncAll(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "workload synced"})
}

const responseDateFormat = "02-01-2006"

// formatTimePtr mengubah *time.Time menjadi *string dengan format yang ditentukan.
func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(responseDateFormat)
	return &formatted
}

// formatEfficiency formats a float64 pointer into a percentage string.
func formatEfficiency(val *float64) *string {
	if val == nil {
		return nil
	}
	s := fmt.Sprintf("%.2f%%", *val)
	return &s
}

// formatRemainingHours formats hours into a "D hari, H jam" string.
func formatRemainingHours(hours *float64) *string {
	if hours == nil {
		return nil
	}
	h := *hours
	sign := ""
	if h < 0 {
		sign = "-"
		h = -h
	}
	days := int(h / 24)
	remainingHours := int(math.Mod(h, 24))
	s := fmt.Sprintf("%s%d hari, %d jam", sign, days, remainingHours)
	return &s
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

	// 1. Ambil data asli dari service
	originalResponse, err := h.workloadSvc.GetTasksByRangeGrouped(c.Request.Context(), startDate, endDate, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Transformasi data ke format respons yang diinginkan
	responseAssignees := make([]AssigneeWithTasks, len(originalResponse.Assignees))
	for i, originalAssignee := range originalResponse.Assignees {
		formattedTasks := make([]TaskInResponse, len(originalAssignee.Tasks))
		for j, originalTask := range originalAssignee.Tasks {
			formattedTasks[j] = TaskInResponse{
				ID:                originalTask.ID,
				Name:              originalTask.Name,
				Description:       originalTask.Description,
				StatusID:          originalTask.StatusID,
				StatusName:        originalTask.StatusName,
				StatusType:        originalTask.StatusType,
				ProjectName:       originalTask.ProjectName,
				TimeEstimateHours: originalTask.TimeEstimateHours,
				TimeSpentHours:    originalTask.TimeSpentHours,
				Category:          originalTask.Category,
				StartDate:         formatTimePtr(originalTask.StartDate),
				DueDate:           formatTimePtr(originalTask.DueDate),
				DateDone:          formatTimePtr(originalTask.DateDone),
				DateClosed:        formatTimePtr(originalTask.DateClosed),
				TimeEfficiencyPercentage: formatEfficiency(originalTask.TimeEfficiencyPercentage),
				RemainingTimeHours:         originalTask.RemainingTimeHours,
				RemainingTimeFormatted:     formatRemainingHours(originalTask.RemainingTimeHours),
			}
		}

		// --- START: Kalkulasi Persentase Tepat Waktu ---
		var onTimeCount int
		var completedCount int
		for _, task := range originalAssignee.Tasks {
			if task.DateDone != nil {
				completedCount++
				if task.DueDate != nil && !task.DateDone.After(*task.DueDate) { // Cek jika donedate <= duedate
					onTimeCount++
				}
			}
		}

		var onTimePercentage *string
		if completedCount > 0 {
			percentage := (float64(onTimeCount) / float64(completedCount)) * 100
			onTimePercentage = formatEfficiency(&percentage) // Menggunakan kembali fungsi format yang ada
		}
		// --- END: Kalkulasi ---

		// --- START: Kalkulasi Actual Work Hours ---
		var totalActualWorkHours float64
		for _, task := range originalAssignee.Tasks {
			// Hitung hanya untuk tugas yang sudah selesai dan memiliki tanggal yang valid
			if task.StartDate != nil && task.DateDone != nil {
				totalActualWorkHours += task.DateDone.Sub(*task.StartDate).Hours()
			}
		}
		// --- END: Kalkulasi ---

		// Salin field dari originalAssignee ke responseAssignees[i] secara manual
		responseAssignees[i] = AssigneeWithTasks{
			ClickupID:          originalAssignee.ClickUpID,
			Username:           originalAssignee.Username,
			Email:              originalAssignee.Email,
			Name:               originalAssignee.Name,
			TotalSpentHours:    originalAssignee.TotalSpentHours,
			ExpectedHours:      originalAssignee.ExpectedHours, 
			TotalTasks:         originalAssignee.TotalTasks,
			ActualWorkHours:    math.Round(totalActualWorkHours),
			TotalUpcomingHours: originalAssignee.TotalUpcomingHours,
			OnTimeCompletionPercentage: onTimePercentage,
			Tasks:              formattedTasks, 
		}
	}

	response := gin.H{
		"count":     originalResponse.Count,
		"assignees": responseAssignees,
	}

	c.JSON(http.StatusOK, response)
}