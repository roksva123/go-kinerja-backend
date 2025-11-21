package handlers

import (
    "net/http"

    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/roksva123/go-kinerja-backend/internal/model"
    "github.com/roksva123/go-kinerja-backend/internal/service"
    "github.com/roksva123/go-kinerja-backend/internal/utils"
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
        "status":       "ok",
        "synced_tasks": count,
    })
}

func (h *ClickUpHandler) SyncTeam(c *gin.Context) {
    err := h.Click.SyncTeam(c)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "status":  "ok",
        "message": "spaces synced as teams",
    })
}

func (h *ClickUpHandler) GetTeams(c *gin.Context) {
    teams, err := h.Click.GetTeams(c)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // convert ke response format
    var resp []model.TeamResponse
    for _, t := range teams {
        resp = append(resp, utils.ConvertTeamToResponse(t))
    }

    c.JSON(200, gin.H{"teams": resp})
}

func (h *ClickUpHandler) GetMembers(c *gin.Context) {
    members, err := h.Click.GetMembers(c)
    
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // memberResponse := []model.User

    for i := range members {
        members[i].Photo = "/images/" + fmt.Sprintf("%d", members[i].ID) + ".jpg"
    }
    c.JSON(200, members)
}


func (h *ClickUpHandler) GetTasks(c *gin.Context) {
    tasks, err := h.Click.GetTasks(c)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "tasks": tasks,
    })
}
