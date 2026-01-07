package repository

import (
	"encoding/json"
	"fmt"
	"pt-besq-core/internal/database"
	"pt-besq-core/internal/entity" // Import Entity
	"time"
)

type InstanceRepository struct{}

func NewInstanceRepository() *InstanceRepository {
	return &InstanceRepository{}
}

// --- BAGIAN 1: VALIDATION & SAVING ---

// ValidateInput memvalidasi data JSON dari user
// Perhatikan: parameter fields menggunakan []entity.FieldDef
func (r *InstanceRepository) ValidateInput(input map[string]interface{}, fields []entity.FieldDef) error {
	for _, field := range fields {
		val, exists := input[field.Key]

		if field.IsRequired && (!exists || val == nil || val == "") {
			return fmt.Errorf("kolom '%s' wajib diisi", field.Label)
		}

		if exists && val != nil && val != "" {
			switch field.Type {
			case "number":
				if _, ok := val.(float64); !ok {
					return fmt.Errorf("kolom '%s' harus berupa angka", field.Label)
				}
			case "text", "date":
				if _, ok := val.(string); !ok {
					return fmt.Errorf("kolom '%s' harus berupa teks", field.Label)
				}
			}
		}
	}
	return nil
}

// SaveInstance menyimpan data baru
func (r *InstanceRepository) SaveInstance(workflowID, templateID int, dataJSON []byte) (int64, error) {
	query := `INSERT INTO process_instances (workflow_id, template_id, status, data_payload, created_at) 
	          VALUES (?, ?, 'draft', ?, NOW())`
	
	res, err := database.DB.Exec(query, workflowID, templateID, dataJSON)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// --- BAGIAN 2: STRUCTS & DATA RETRIEVAL ---

type InstanceLog struct {
	ID           int64           `db:"id" json:"id"`
	WorkflowName string          `db:"workflow_name" json:"workflow_name"`
	TemplateName string          `db:"template_name" json:"template_name"`
	Status       string          `db:"status" json:"status"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
	DataPayload  json.RawMessage `db:"data_payload" json:"data_payload"`
}

type DailyStat struct {
	TemplateName string `db:"template_name"`
	Count        int    `db:"total"`
}

// GetHistory mengambil list data dengan pagination
func (r *InstanceRepository) GetHistory(limit, offset int, templateID int, dateStr string) ([]InstanceLog, error) {
	var logs []InstanceLog
	query := `
		SELECT i.id, w.name as workflow_name, t.name as template_name, 
		       i.status, i.created_at, i.data_payload
		FROM process_instances i
		JOIN workflows w ON i.workflow_id = w.id
		JOIN process_templates t ON i.template_id = t.id
		WHERE 1=1 
	`
	var args []interface{}

	if templateID > 0 {
		query += " AND i.template_id = ? "
		args = append(args, templateID)
	}
	if dateStr != "" {
		query += " AND DATE(i.created_at) = ? "
		args = append(args, dateStr)
	}

	query += " ORDER BY i.created_at DESC LIMIT ? OFFSET ? "
	args = append(args, limit, offset)

	err := database.DB.Select(&logs, query, args...)
	return logs, err
}

// CountHistory menghitung total baris data
func (r *InstanceRepository) CountHistory(templateID int, dateStr string) (int, error) {
	var total int
	query := "SELECT COUNT(*) FROM process_instances i WHERE 1=1 "
	var args []interface{}

	if templateID > 0 {
		query += " AND i.template_id = ? "
		args = append(args, templateID)
	}
	if dateStr != "" {
		query += " AND DATE(i.created_at) = ? "
		args = append(args, dateStr)
	}

	err := database.DB.Get(&total, query, args...)
	return total, err
}

// GetDailyStats menghitung statistik harian
func (r *InstanceRepository) GetDailyStats() ([]DailyStat, error) {
	var stats []DailyStat
	query := `
		SELECT t.name AS template_name, COUNT(i.id) AS total
		FROM process_instances i
		JOIN process_templates t ON i.template_id = t.id
		WHERE DATE(i.created_at) = CURRENT_DATE
		GROUP BY t.name
	`
	err := database.DB.Select(&stats, query)
	return stats, err
}