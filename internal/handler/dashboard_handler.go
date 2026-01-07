package handler

import (
	"net/http"
	"pt-besq-core/internal/repository"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	Repo *repository.InstanceRepository
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		Repo: repository.NewInstanceRepository(), // Reuse repo instance
	}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	// 1. Ambil data statistik dari DB
	stats, err := h.Repo.GetDailyStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal hitung statistik: " + err.Error()})
		return
	}

	// 2. Hitung Total Keseluruhan (Sum)
	totalToday := 0
	for _, s := range stats {
		totalToday += s.Count
	}

	// 3. Kirim Response JSON yang rapi
	c.JSON(http.StatusOK, gin.H{
		"date":        "today",
		"total_today": totalToday, // Angka besar untuk Dashboard
		"breakdown":   stats,      // Data untuk Grafik Pie/Bar Chart
	})
}