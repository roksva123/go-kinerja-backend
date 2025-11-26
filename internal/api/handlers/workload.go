package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roksva123/go-kinerja-backend/internal/service"
)

type WorkloadHandler struct {
	Work *service.WorkloadService
	Click *service.ClickUpService
}

func NewWorkloadHandler(w *service.WorkloadService, c *service.ClickUpService) *WorkloadHandler {
	return &WorkloadHandler{Work: w, Click: c}
}

func (h *WorkloadHandler) GetWorkload(c *gin.Context) {
	startS := c.Query("start")
	endS := c.Query("end")
	posisi := c.Query("posisi")
	source := c.Query("source")
	name := c.Query("name")

	if startS == "" || endS == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query params required"})
		return
	}
	start, err := time.Parse("2006-01-02", startS)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
		return
	}
	end, err := time.Parse("2006-01-02", endS)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
		return
	}

	resp, err := h.Work.BuildWorkload(c.Request.Context(), start, end, posisi, source, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WorkloadHandler) SyncAll(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.Click.SyncTeam(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync team: " + err.Error()})
		return
	}
	// sync members
	if err := h.Click.SyncMembers(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync members: " + err.Error()})
		return
	}
	// sync tasks
	count, err := h.Click.SyncTasks(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync tasks: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "synced_tasks": count})
}
