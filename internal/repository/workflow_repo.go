package repository

import (
	"pt-besq-core/internal/database"
	"pt-besq-core/internal/entity"
)

type WorkflowRepository struct{}

func NewWorkflowRepository() *WorkflowRepository {
	return &WorkflowRepository{}
}

func (r *WorkflowRepository) GetAll() ([]entity.Workflow, error) {
	var workflows []entity.Workflow
	// Ambil semua workflow (default kosong jika null)
	query := "SELECT id, name, canvas_config, is_active, created_at FROM workflows"
	err := database.DB.Select(&workflows, query)
	return workflows, err
}

func (r *WorkflowRepository) Create(wf entity.Workflow) (int64, error) {
	// Pastikan canvas_config defaultnya valid JSON object '{}' jika kosong
	if len(wf.CanvasConfig) == 0 {
		wf.CanvasConfig = []byte("{}")
	}
	query := `INSERT INTO workflows (name, canvas_config, is_active) VALUES (?, ?, ?)`
	res, err := database.DB.Exec(query, wf.Name, wf.CanvasConfig, wf.IsActive)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *WorkflowRepository) UpdateLayout(id int, configJSON string) error {
	query := `UPDATE workflows SET canvas_config = ? WHERE id = ?`
	_, err := database.DB.Exec(query, configJSON, id)
	return err
}