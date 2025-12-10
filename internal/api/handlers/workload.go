package handlers

// import (
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/roksva123/go-kinerja-backend/internal/service"
// )

// type WorkloadHandler struct {
// 	workloadSvc *service.WorkloadService
// 	clickupSvc  *service.ClickUpService
// }

// func NewWorkloadHandler(w *service.WorkloadService, c *service.ClickUpService) *WorkloadHandler {
// 	return &WorkloadHandler{workloadSvc: w, clickupSvc: c}
// }

// func (h *WorkloadHandler) GetWorkload(c *gin.Context) {
// 	startS := c.Query("start")
// 	endS := c.Query("end")
// 	posisi := c.Query("posisi")
// 	source := c.Query("source")
// 	name := c.Query("name")

// 	if startS == "" || endS == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query params required"})
// 		return
// 	}
// 	start, err := time.Parse("2006-01-02", startS)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
// 		return
// 	}
// 	end, err := time.Parse("2006-01-02", endS)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
// 		return
// 	}

	// resp, err := h.Work.BuildWorkload(c.Request.Context(), start, end, posisi, source, name)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, resp)
// }

// func (h *WorkloadHandler) GetTasksSummary(c *gin.Context) {
// 	startDateStr := c.Query("start_date")
// 	endDateStr := c.Query("end_date")
// 	name := c.Query("name")
// 	email := c.Query("email")

// 	if startDateStr == "" || endDateStr == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
// 		return
// 	}

// 	startDate, err := time.Parse("2006-01-02", startDateStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
// 		return
// 	}

// 	endDate, err := time.Parse("2006-01-02", endDateStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
// 		return
// 	}

// 	summary, err := h.workloadSvc.GetTasksSummary(c.Request.Context(), startDate, endDate, name, email)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, summary)
// }
