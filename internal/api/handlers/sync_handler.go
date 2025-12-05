package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roksva123/go-kinerja-backend/internal/service"
)

type SyncHandler struct {
	ClickUpService *service.ClickUpService
}

func NewSyncHandler(s *service.ClickUpService) *SyncHandler {
	return &SyncHandler{ClickUpService: s}
}

// SyncSpacesFoldersAndListsHandler adalah endpoint untuk trigger sinkronisasi space, folder, dan list.
func (h *SyncHandler) SyncSpacesFoldersAndListsHandler(c *gin.Context) {
	log.Println("--- API TRIGGER: Syncing Spaces, Folders, and Lists ---")
	err := h.ClickUpService.SyncSpacesAndFolders(c.Request.Context())
	if err != nil {
		log.Printf("ERROR from SyncSpacesAndFolders service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sync for spaces, folders, and lists completed successfully"})
	log.Println("--- API TRIGGER: Sync finished successfully ---")
}

func (h *SyncHandler) GetListsHandler(c *gin.Context) {
	lists, err := h.ClickUpService.GetLists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lists)
}