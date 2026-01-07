package handler

import (
	"net/http"
	"pt-besq-core/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	Repo *repository.AuditRepository
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{
		Repo: repository.NewAuditRepository(),
	}
}

// GetLogs menampilkan 100 aktivitas terakhir
func (h *AuditHandler) GetLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	logs, err := h.Repo.GetAllLogs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil log audit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}