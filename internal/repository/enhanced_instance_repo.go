package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"pt-besq-core/internal/database"
	"time"
)

// InstanceWithDetails represents a process instance with all related data
type InstanceWithDetails struct {
	ID               int64           `db:"id" json:"id"`
	TemplateID       int             `db:"template_id" json:"template_id"`
	TemplateName     string          `db:"template_name" json:"template_name"`
	WorkflowID       int             `db:"workflow_id" json:"workflow_id"`
	WorkflowName     string          `db:"workflow_name" json:"workflow_name"`
	BatchNumber      string          `db:"batch_number" json:"batch_number"`
	Status           string          `db:"status" json:"status"`
	Priority         string          `db:"priority" json:"priority"`
	DataPayload      json.RawMessage `db:"data_payload" json:"data_payload"`
	StartTime        *time.Time      `db:"start_time" json:"start_time,omitempty"`
	EndTime          *time.Time      `db:"end_time" json:"end_time,omitempty"`
	DurationMinutes  *int            `db:"duration_minutes" json:"duration_minutes,omitempty"`
	Notes            string          `db:"notes" json:"notes,omitempty"`
	CreatedBy        int             `db:"created_by" json:"created_by"`
	CreatedByName    string          `db:"created_by_name" json:"created_by_name"`
	ApprovedBy       *int            `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedByName   *string         `db:"approved_by_name" json:"approved_by_name,omitempty"`
	ApprovedAt       *time.Time      `db:"approved_at" json:"approved_at,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`
}

// InstanceHistoryEntry represents a change in an instance
type InstanceHistoryEntry struct {
	ID            int64           `db:"id" json:"id"`
	InstanceID    int64           `db:"instance_id" json:"instance_id"`
	Action        string          `db:"action" json:"action"`
	OldValue      json.RawMessage `db:"old_value" json:"old_value,omitempty"`
	NewValue      json.RawMessage `db:"new_value" json:"new_value,omitempty"`
	ChangedBy     int             `db:"changed_by" json:"changed_by"`
	ChangedByName string          `db:"changed_by_name" json:"changed_by_name"`
	Comment       string          `db:"comment" json:"comment,omitempty"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
}

// ProductionStats represents production statistics
type ProductionStats struct {
	TotalInstances      int            `json:"total_instances"`
	CompletedInstances  int            `json:"completed_instances"`
	InProgressInstances int            `json:"in_progress_instances"`
	RejectedInstances   int            `json:"rejected_instances"`
	AverageDuration     float64        `json:"average_duration_minutes"`
	ByTemplate          []TemplateStat `json:"by_template"`
	ByStatus            []StatusStat   `json:"by_status"`
	ByPriority          []PriorityStat `json:"by_priority"`
}

type TemplateStat struct {
	TemplateName string  `db:"template_name" json:"template_name"`
	Count        int     `db:"count" json:"count"`
	Percentage   float64 `json:"percentage"`
}

type StatusStat struct {
	Status     string  `db:"status" json:"status"`
	Count      int     `db:"count" json:"count"`
	Percentage float64 `json:"percentage"`
}

type PriorityStat struct {
	Priority   string  `db:"priority" json:"priority"`
	Count      int     `db:"count" json:"count"`
	Percentage float64 `json:"percentage"`
}

// EnhancedInstanceRepository provides advanced instance operations
type EnhancedInstanceRepository struct{}

func NewEnhancedInstanceRepository() *EnhancedInstanceRepository {
	return &EnhancedInstanceRepository{}
}

// GetWithDetails retrieves instance with all related information
func (r *EnhancedInstanceRepository) GetWithDetails(instanceID int64) (*InstanceWithDetails, error) {
	var instance InstanceWithDetails
	query := `
		SELECT 
			i.id, i.template_id, t.name as template_name,
			i.workflow_id, w.name as workflow_name,
			i.batch_number, i.status, i.priority, i.data_payload,
			i.start_time, i.end_time, i.duration_minutes, i.notes,
			i.created_by, u1.full_name as created_by_name,
			i.approved_by, u2.full_name as approved_by_name, i.approved_at,
			i.created_at, i.updated_at
		FROM process_instances i
		JOIN process_templates t ON i.template_id = t.id
		JOIN workflows w ON i.workflow_id = w.id
		LEFT JOIN users u1 ON i.created_by = u1.id
		LEFT JOIN users u2 ON i.approved_by = u2.id
		WHERE i.id = ?
	`
	err := database.DB.Get(&instance, query, instanceID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("instance not found")
	}
	return &instance, err
}

// UpdateStatus changes the status of an instance and logs the change
func (r *EnhancedInstanceRepository) UpdateStatus(instanceID int64, newStatus string, changedBy int, comment string) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current status
	var oldStatus string
	err = tx.Get(&oldStatus, "SELECT status FROM process_instances WHERE id = ?", instanceID)
	if err != nil {
		return err
	}

	// Update status
	_, err = tx.Exec(`
		UPDATE process_instances 
		SET status = ?, updated_at = NOW() 
		WHERE id = ?
	`, newStatus, instanceID)
	if err != nil {
		return err
	}

	// Log the change
	oldVal, _ := json.Marshal(map[string]string{"status": oldStatus})
	newVal, _ := json.Marshal(map[string]string{"status": newStatus})

	_, err = tx.Exec(`
		INSERT INTO instance_history (instance_id, action, old_value, new_value, changed_by, comment, created_at)
		VALUES (?, 'status_changed', ?, ?, ?, ?, NOW())
	`, instanceID, oldVal, newVal, changedBy, comment)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ApproveInstance approves an instance
func (r *EnhancedInstanceRepository) ApproveInstance(instanceID int64, approvedBy int, comment string) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update instance
	_, err = tx.Exec(`
		UPDATE process_instances 
		SET status = 'completed', approved_by = ?, approved_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`, approvedBy, instanceID)
	if err != nil {
		return err
	}

	// Log the approval
	newVal, _ := json.Marshal(map[string]interface{}{
		"status":      "completed",
		"approved_by": approvedBy,
	})

	_, err = tx.Exec(`
		INSERT INTO instance_history (instance_id, action, new_value, changed_by, comment, created_at)
		VALUES (?, 'approved', ?, ?, ?, NOW())
	`, instanceID, newVal, approvedBy, comment)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RejectInstance rejects an instance
func (r *EnhancedInstanceRepository) RejectInstance(instanceID int64, rejectedBy int, reason string) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update instance
	_, err = tx.Exec(`
		UPDATE process_instances 
		SET status = 'rejected', notes = ?, updated_at = NOW()
		WHERE id = ?
	`, reason, instanceID)
	if err != nil {
		return err
	}

	// Log the rejection
	newVal, _ := json.Marshal(map[string]interface{}{
		"status": "rejected",
		"reason": reason,
	})

	_, err = tx.Exec(`
		INSERT INTO instance_history (instance_id, action, new_value, changed_by, comment, created_at)
		VALUES (?, 'rejected', ?, ?, ?, NOW())
	`, instanceID, newVal, rejectedBy, reason)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetHistory retrieves the change history of an instance
func (r *EnhancedInstanceRepository) GetHistory(instanceID int64) ([]InstanceHistoryEntry, error) {
	var history []InstanceHistoryEntry
	query := `
		SELECT 
			h.id, h.instance_id, h.action, h.old_value, h.new_value,
			h.changed_by, COALESCE(u.full_name, 'System') as changed_by_name, h.comment, h.created_at
		FROM instance_history h
		LEFT JOIN users u ON h.changed_by = u.id
		WHERE h.instance_id = ?
		ORDER BY h.created_at DESC
	`
	err := database.DB.Select(&history, query, instanceID)
	return history, err
}

// GetProductionStats retrieves comprehensive production statistics
func (r *EnhancedInstanceRepository) GetProductionStats(startDate, endDate *time.Time) (*ProductionStats, error) {
	stats := &ProductionStats{}

	// Build date filter
	dateFilter := ""
	args := []interface{}{}
	if startDate != nil && endDate != nil {
		dateFilter = "WHERE i.created_at BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	// Get total counts - FIXED: Added proper alias and handle NULL values
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN i.status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN i.status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN i.status = 'rejected' THEN 1 ELSE 0 END) as rejected,
			COALESCE(AVG(CASE WHEN i.duration_minutes IS NOT NULL THEN i.duration_minutes ELSE 0 END), 0) as avg_duration
		FROM process_instances i
		%s
	`, dateFilter)

	err := database.DB.QueryRow(query, args...).Scan(
		&stats.TotalInstances,
		&stats.CompletedInstances,
		&stats.InProgressInstances,
		&stats.RejectedInstances,
		&stats.AverageDuration,
	)
	if err != nil {
		return nil, err
	}

	// Get by template
	query = fmt.Sprintf(`
		SELECT t.name as template_name, COUNT(*) as count
		FROM process_instances i
		JOIN process_templates t ON i.template_id = t.id
		%s
		GROUP BY t.name
		ORDER BY count DESC
	`, dateFilter)

	err = database.DB.Select(&stats.ByTemplate, query, args...)
	if err != nil {
		return nil, err
	}

	// Calculate percentages
	for i := range stats.ByTemplate {
		if stats.TotalInstances > 0 {
			stats.ByTemplate[i].Percentage = float64(stats.ByTemplate[i].Count) / float64(stats.TotalInstances) * 100
		}
	}

	// Get by status
	query = fmt.Sprintf(`
		SELECT i.status, COUNT(*) as count
		FROM process_instances i
		%s
		GROUP BY i.status
		ORDER BY count DESC
	`, dateFilter)

	err = database.DB.Select(&stats.ByStatus, query, args...)
	if err != nil {
		return nil, err
	}

	for i := range stats.ByStatus {
		if stats.TotalInstances > 0 {
			stats.ByStatus[i].Percentage = float64(stats.ByStatus[i].Count) / float64(stats.TotalInstances) * 100
		}
	}

	// Get by priority
	query = fmt.Sprintf(`
		SELECT i.priority, COUNT(*) as count
		FROM process_instances i
		%s
		GROUP BY i.priority
		ORDER BY count DESC
	`, dateFilter)

	err = database.DB.Select(&stats.ByPriority, query, args...)
	if err != nil {
		return nil, err
	}

	for i := range stats.ByPriority {
		if stats.TotalInstances > 0 {
			stats.ByPriority[i].Percentage = float64(stats.ByPriority[i].Count) / float64(stats.TotalInstances) * 100
		}
	}

	return stats, nil
}

// SearchInstances provides advanced search functionality
func (r *EnhancedInstanceRepository) SearchInstances(filters map[string]interface{}, page, limit int) ([]InstanceWithDetails, int, error) {
	var instances []InstanceWithDetails

	query := `
		SELECT 
			i.id, i.template_id, t.name as template_name,
			i.workflow_id, w.name as workflow_name,
			i.batch_number, i.status, i.priority, i.data_payload,
			i.start_time, i.end_time, i.duration_minutes, i.notes,
			i.created_by, COALESCE(u1.full_name, 'Unknown') as created_by_name,
			i.approved_by, u2.full_name as approved_by_name, i.approved_at,
			i.created_at, i.updated_at
		FROM process_instances i
		JOIN process_templates t ON i.template_id = t.id
		JOIN workflows w ON i.workflow_id = w.id
		LEFT JOIN users u1 ON i.created_by = u1.id
		LEFT JOIN users u2 ON i.approved_by = u2.id
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM process_instances i WHERE 1=1"

	args := []interface{}{}

	// Apply filters
	if templateID, ok := filters["template_id"].(int); ok && templateID > 0 {
		query += " AND i.template_id = ?"
		countQuery += " AND i.template_id = ?"
		args = append(args, templateID)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += " AND i.status = ?"
		countQuery += " AND i.status = ?"
		args = append(args, status)
	}

	if priority, ok := filters["priority"].(string); ok && priority != "" {
		query += " AND i.priority = ?"
		countQuery += " AND i.priority = ?"
		args = append(args, priority)
	}

	if batchNumber, ok := filters["batch_number"].(string); ok && batchNumber != "" {
		query += " AND i.batch_number LIKE ?"
		countQuery += " AND i.batch_number LIKE ?"
		args = append(args, "%"+batchNumber+"%")
	}

	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		query += " AND DATE(i.created_at) >= ?"
		countQuery += " AND DATE(i.created_at) >= ?"
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		query += " AND DATE(i.created_at) <= ?"
		countQuery += " AND DATE(i.created_at) <= ?"
		args = append(args, endDate)
	}

	// Get total count
	var total int
	err := database.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	offset := (page - 1) * limit
	query += " ORDER BY i.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Execute query
	err = database.DB.Select(&instances, query, args...)
	return instances, total, err
}