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

func (h *ClickUpHandler) GetSpaces(c *gin.Context) {
	spaces, err := h.Click.GetSpaces(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"spaces": spaces})
}

func (h *ClickUpHandler) GetFullSyncFiltered(c *gin.Context) {
	var filter model.FullSyncFilter

	filter.Username = c.Query("username")
	filter.Email = c.Query("email")
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