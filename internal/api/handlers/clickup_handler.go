package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/service"
)

// CLICKUP HANDLER

type ClickUpHandler struct {
	Click *service.ClickUpService
}

func NewClickUpHandler(click *service.ClickUpService) *ClickUpHandler {
	return &ClickUpHandler{Click: click}
}

func (h *ClickUpHandler) SyncTeam(c *gin.Context) {
	if err := h.Click.SyncTeam(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "teams synced"})
}

func (h *ClickUpHandler) SyncMembers(c *gin.Context) {
	if err := h.Click.SyncMembers(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "members synced"})
}

func (h *ClickUpHandler) SyncTasks(c *gin.Context) {
	n, err := h.Click.SyncTasks(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tasks synced", "count": n})
}

func (h *ClickUpHandler) SyncAll(c *gin.Context) {
	err := h.Click.AllSync(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "full sync completed successfully"})
}

func (h *ClickUpHandler) GetTasks(c *gin.Context) {
	tasks, err := h.Click.GetTasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *ClickUpHandler) GetMembers(c *gin.Context) {
	users, err := h.Click.GetMembers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *ClickUpHandler) GetTeams(c *gin.Context) {
	teams, err := h.Click.GetTeams(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (h *ClickUpHandler) GetFullSync(c *gin.Context) {
	data, err := h.Click.FullSync(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *ClickUpHandler) GetFullSyncFiltered(c *gin.Context) {
	var filter model.FullSyncFilter

	filter.Role = c.Query("role")
	filter.Range = c.Query("range")

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr != "" && endStr != "" {
		layout := "2006-01-02"

		if sd, err := time.Parse(layout, startStr); err == nil {
			ms := sd.UnixNano() / 1e6
			filter.StartDate = &ms
		}
		if ed, err := time.Parse(layout, endStr); err == nil {
			ed = ed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			ms := ed.UnixNano() / 1e6
			filter.EndDate = &ms
		}
	}

	out, err := h.Click.FullSyncFiltered(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": out})
}

func (h *ClickUpHandler) FullSync(c *gin.Context) {
	var filter model.FullSyncFilter

	filter.Role = c.Query("role")
	filter.Username = c.Query("username")

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr != "" && endStr != "" {
		layout := "2006-01-02"

		if sd, err := time.Parse(layout, startStr); err == nil {
			ms := sd.UnixNano() / 1e6
			filter.StartDate = &ms
		}
		if ed, err := time.Parse(layout, endStr); err == nil {
			ed = ed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			ms := ed.UnixNano() / 1e6
			filter.EndDate = &ms
		}
	}

	out, err := h.Click.FullSyncFlow(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *ClickUpHandler) GetFullData(c *gin.Context) {
	var filter model.FullSyncFilter

	filter.Role = c.Query("role")
	filter.Username = c.Query("username")

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr != "" && endStr != "" {
		layout := "2006-01-02"

		if sd, err := time.Parse(layout, startStr); err == nil {
			ms := sd.UnixNano() / 1e6
			filter.StartDate = &ms
		}
		if ed, err := time.Parse(layout, endStr); err == nil {
			ed = ed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			ms := ed.UnixNano() / 1e6
			filter.EndDate = &ms
		}
	}

	out, err := h.Click.Repo.GetFullDataFiltered(
		c.Request.Context(),
		filter.StartDate,
		filter.EndDate,
		filter.Role,
		filter.Username,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

// WORKLOAD HANDLER 

type WorkloadHandler struct {
	Svc      *service.WorkloadService
	ClickSvc *service.ClickUpService
}

func NewWorkloadHandler(svc *service.WorkloadService, clickSvc *service.ClickUpService) *WorkloadHandler {
	return &WorkloadHandler{
		Svc:      svc,
		ClickSvc: clickSvc,
	}
}

func (h *WorkloadHandler) GetWorkload(c *gin.Context) {
	startStr := c.Query("start") 
	endStr := c.Query("end")     

	layout := "2006-01-02"
	start, err := time.Parse(layout, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use YYYY-MM-DD"})
		return
	}
	end, err := time.Parse(layout, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use YYYY-MM-DD"})
		return
	}

	// Convert to Unix milliseconds
	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	users, err := h.ClickSvc.GetWorkload(c.Request.Context(), startMs, endMs)
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
	err := h.Svc.SyncAll(c.Request.Context())
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

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse(layout, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
		return
	}

	assigneesMap, tasksByAssignee, err := h.Svc.GetTasksByRangeGroupedByAssignee(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responseAssignees []AssigneeWithTasks
	for id, assignee := range assigneesMap {
		tasksForResponse := make([]TaskInResponse, len(tasksByAssignee[id]))
		for i, taskDetail := range tasksByAssignee[id] {
			tasksForResponse[i] = TaskInResponse{
				ID:                taskDetail.ID,
				Name:              taskDetail.Name,
				Description:       taskDetail.Description,
				TextContent:       taskDetail.TextContent,
				StatusName:        taskDetail.StatusName,
				StartDate:         taskDetail.StartDate,
				DueDate:           taskDetail.DueDate,
				DateDone:          taskDetail.DateDone,
				TimeEstimateHours: taskDetail.TimeEstimateHours,
				TimeSpentHours:    taskDetail.TimeSpentHours,
			}
		}

		responseAssignees = append(responseAssignees, AssigneeWithTasks{
			ClickupID: int(assignee.ClickUpID),
			Username:  assignee.Username,
			Email:     assignee.Email,
			Name:      assignee.Name,
			Tasks:     tasksForResponse,
		})
	}

	finalResponse := TasksByAssigneeResponse{
		Count:     len(responseAssignees),
		Assignees: responseAssignees,
	}

	c.JSON(http.StatusOK, finalResponse)
}