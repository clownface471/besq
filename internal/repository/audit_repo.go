package repository

import (
	"pt-besq-core/internal/database"
	"time"
)

// Struct Entity untuk Log
type ActivityLog struct {
	ID         int       `db:"id" json:"id"`
	UserID     *int      `db:"user_id" json:"user_id"` // Pointer karena bisa NULL
	Username   string    `db:"username" json:"username"`
	Method     string    `db:"method" json:"method"`
	Path       string    `db:"path" json:"path"`
	IPAddress  string    `db:"ip_address" json:"ip_address"`
	StatusCode int       `db:"status_code" json:"status_code"`
	UserAgent  string    `db:"user_agent" json:"user_agent"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type AuditRepository struct{}

func NewAuditRepository() *AuditRepository {
	return &AuditRepository{}
}

// LogActivity menyimpan jejak aktivitas baru
func (r *AuditRepository) LogActivity(log ActivityLog) error {
	query := `
		INSERT INTO activity_logs (user_id, username, method, path, ip_address, status_code, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := database.DB.Exec(query,
		log.UserID, log.Username, log.Method, log.Path, log.IPAddress, log.StatusCode, log.UserAgent,
	)
	return err
}

// GetAllLogs mengambil daftar log (Khusus Admin)
func (r *AuditRepository) GetAllLogs(limit int) ([]ActivityLog, error) {
	var logs []ActivityLog
	// Ambil data terbaru dulu
	query := `SELECT * FROM activity_logs ORDER BY created_at DESC LIMIT ?`
	err := database.DB.Select(&logs, query, limit)
	return logs, err
}