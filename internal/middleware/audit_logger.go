package middleware

import (
	"pt-besq-core/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogger adalah middleware perekam jejak
func AuditLogger() gin.HandlerFunc {
	// Kita inisialisasi repo di sini
	repo := repository.NewAuditRepository()

	return func(c *gin.Context) {
		// 1. Catat waktu mulai
		startTime := time.Now()

		// 2. Biarkan request diproses oleh handler berikutnya (Next)
		c.Next()

		// 3. SETELAH request selesai, kita kumpulkan data
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// Ambil data User (jika ada, dari AuthMiddleware)
		var userID *int
		username := "Guest"

		if val, exists := c.Get("user_id"); exists {
			// Konversi float64 (default JSON number) ke int
			if idFloat, ok := val.(float64); ok {
				idInt := int(idFloat)
				userID = &idInt
			} else if idInt, ok := val.(int); ok { // Fallback jika sudah int
				userID = &idInt
			}
		}

		if val, exists := c.Get("username"); exists {
			username = val.(string)
		}

		// 4. Susun Laporan Log
		logEntry := repository.ActivityLog{
			UserID:     userID,
			Username:   username,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IPAddress:  c.ClientIP(),
			StatusCode: c.Writer.Status(),
			UserAgent:  c.Request.UserAgent(),
		}

		// 5. Simpan ke Database (Jalankan di Goroutine agar tidak memperlambat respon ke user)
		go func(l repository.ActivityLog, lat time.Duration) {
			_ = repo.LogActivity(l)
			// (Opsional) Print ke terminal juga biar kelihatan saat dev
			// fmt.Printf("[AUDIT] %s | %d | %s | %v\n", l.Path, l.StatusCode, l.Username, lat)
		}(logEntry, latency)
	}
}