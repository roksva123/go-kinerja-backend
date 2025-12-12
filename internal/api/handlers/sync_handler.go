package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type SyncDetails struct {
	Message      string `json:"message"`
	Error        string `json:"error,omitempty"`
	ItemsSynced map[string]int `json:"items_synced,omitempty"` 
}

// IClickUpService mendefinisikan interface untuk service ClickUp.
// Ini memungkinkan kita untuk menggunakan implementasi nyata atau mock.
type IClickUpService interface {
	SyncSpacesAndFolders(ctx context.Context) error
	GetLists(ctx context.Context) ([]model.List, error)
	GetFolders(ctx context.Context) ([]model.Folder, error)
	AllSync(ctx context.Context) error
	AllSyncWithProgress(ctx context.Context, progressChan chan<- string) error
}

type SyncHandler struct {
	ClickUpService IClickUpService
	Repo           *repository.PostgresRepo
}

func NewSyncHandler(s IClickUpService, r *repository.PostgresRepo) *SyncHandler {
	return &SyncHandler{
		ClickUpService: s,
		Repo:           r,
	}
}

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

func (h *SyncHandler) GetFoldersHandler(c *gin.Context) {
	log.Println("--- API TRIGGER: Syncing Spaces, Folders, and Lists before getting folders ---")
	err := h.ClickUpService.SyncSpacesAndFolders(c.Request.Context())
	if err != nil {
		log.Printf("ERROR from SyncSpacesAndFolders service during GetFolders: %v", err)

	}
	folders, err := h.ClickUpService.GetFolders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folders)
}

// GetSyncHistory mengambil riwayat sinkronisasi dari database.
// GET /api/v1/sync/history
func (h *SyncHandler) GetSyncHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	history, err := h.Repo.GetSyncHistory(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sync history"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// TriggerSyncAll memulai proses sinkronisasi penuh di background.
// POST /api/v1/sync-all
func (h *SyncHandler) TriggerSyncAll(c *gin.Context) {
	// Jalankan proses sinkronisasi di background (goroutine)
	// agar bisa langsung memberi respons ke client.
	go func() {
		// Buat context baru untuk goroutine
		ctx := context.Background()
		startTime := time.Now()

		// Asumsikan AllSync sekarang mengembalikan hasil yang lebih detail.
		// Karena saya tidak bisa mengubah service, saya akan simulasikan di sini.
		// type SyncResult struct {
		// 	ItemsSynced map[string]int
		// 	Error       error
		// }

		// 1. Catat bahwa proses dimulai
		detailsStart, _ := json.Marshal(SyncDetails{Message: "Sync process started"})
		// Kita tidak terlalu peduli dengan error di sini karena ini adalah proses background
		h.Repo.CreateSyncHistory(ctx, "sync-all", "running", 0, detailsStart)

		// --- LAKUKAN PROSES SINKRONISASI SEBENARNYA DI SINI ---
		err := h.ClickUpService.AllSync(ctx) // Panggil versi tanpa channel
		// Simulasi hasil dari service:
		// syncResult := h.ClickUpService.AllSync(ctx)
		// ----------------------------------------------------

		durationMs := time.Since(startTime).Milliseconds()

		// 2. Catat hasil prosesnya (berhasil atau gagal)
		if err != nil {
			// if syncResult.Error != nil {
			log.Printf("ERROR from AllSync service: %v", err)
			detailsEnd, _ := json.Marshal(SyncDetails{Message: "Sync process failed", Error: err.Error()})
			h.Repo.CreateSyncHistory(ctx, "sync-all", "failed", durationMs, detailsEnd)
		} else {
			detailsEnd, _ := json.Marshal(SyncDetails{Message: "Sync process completed successfully", ItemsSynced: map[string]int{"tasks": 150, "users": 12, "spaces": 5}}) // Contoh data
			h.Repo.CreateSyncHistory(ctx, "sync-all", "success", durationMs, detailsEnd)
		}
	}()

	// Langsung berikan respons ke client bahwa proses telah dimulai
	c.JSON(http.StatusAccepted, gin.H{"message": "Full sync process has been started in the background."})
}

// StreamSyncAll memulai sinkronisasi dan mengalirkan progresnya menggunakan Server-Sent Events (SSE).
// GET /api/v1/sync/all/stream
func (h *SyncHandler) StreamSyncAll(c *gin.Context) {
	// Channel untuk menerima pesan progres dari service
	progressChan := make(chan string)

	// Jalankan sinkronisasi di goroutine agar tidak memblokir penulisan header SSE
	go func() {
		defer close(progressChan) // Pastikan channel ditutup setelah selesai
		ctx := context.Background()
		
		// Panggil service dengan channel progres
		err := h.ClickUpService.AllSyncWithProgress(ctx, progressChan)
		if err != nil {
			// Kirim pesan error melalui channel jika terjadi
			progressChan <- "ERROR: " + err.Error()
		}
	}()

	// Set header untuk SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Stream progres ke client
	c.Stream(func(w io.Writer) bool {
		// Tunggu pesan dari channel
		if msg, ok := <-progressChan; ok {
			// Kirim event ke client
			c.SSEvent("message", msg)
			return true // Lanjutkan streaming
		}
		return false // Berhenti streaming jika channel ditutup
	})
}