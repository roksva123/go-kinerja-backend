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
		tasksForAssignee := tasksByAssignee[id]
		totalSpentHours := 0.0
		var tasksInResponse []TaskInResponse

		for _, task := range tasksForAssignee {
			totalSpentHours += task.TimeSpentHours
			tasksInResponse = append(tasksInResponse, TaskInResponse{
				ID:                task.ID,
				Name:              task.Name,
				Description:       task.Description,
				ProjectName:       task.ProjectName,
				StatusName:        task.StatusName,
				StartDate:         task.StartDate,
				DueDate:           task.DueDate,
				DateDone:          task.DateDone,
				TimeEstimateHours: task.TimeEstimateHours,
				TimeSpentHours:    task.TimeSpentHours,
			})
		}

		responseAssignees = append(responseAssignees, AssigneeWithTasks{
			ClickupID:       int(assignee.ClickUpID),
			Username:        assignee.Username,
			Email:           assignee.Email,
			Name:            assignee.Name,
			Tasks:           tasksInResponse,
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