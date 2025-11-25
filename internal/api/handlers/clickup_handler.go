package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roksva123/go-kinerja-backend/internal/service"
    "github.com/roksva123/go-kinerja-backend/internal/model"
)

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
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, data)
}


func (h *ClickUpHandler) GetFullSyncFiltered(c *gin.Context) {
	var filter model.FullSyncFilter

	// role
	filter.Role = c.Query("role")

	// range param
	filter.Range = c.Query("range")

	// custom start/end (priority)
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")
	if startStr != "" && endStr != "" {
		layout := "2006-01-02"
		if sd, err := time.Parse(layout, startStr); err == nil {
			ms := sd.UnixNano() / int64(1e6)
			filter.StartDate = &ms
		}
		if ed, err := time.Parse(layout, endStr); err == nil {
			// end of day -> include full day
			ed = ed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			ms := ed.UnixNano() / int64(1e6)
			filter.EndDate = &ms
		}
	}

	// call service
	out, err := h.Click.FullSyncFiltered(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": out})
}