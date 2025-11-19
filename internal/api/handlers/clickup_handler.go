package handlers

import (
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

func (h *ClickUpHandler) SyncMembers(c *gin.Context) {
    err := h.Click.SyncUsersFromClickUp(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  "ok",
        "message": "members synchronized",
    })
}

func (h *ClickUpHandler) SyncTasks(c *gin.Context) {
    count, err := h.Click.SyncTasks(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
        "synced_tasks": count,
    })
}


func (h *ClickUpHandler) SyncTeam(c *gin.Context) {
    teams, err := h.Click.FetchTeams()
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{
            "error": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "teams": teams,
    })
}
