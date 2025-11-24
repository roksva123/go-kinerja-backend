package handlers

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/roksva123/go-kinerja-backend/internal/service"

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
