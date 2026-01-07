package entity

import (
	"encoding/json"
	"time"
)

// ProcessTemplate: Cetakan dasar (misal: "Mixing", "Oven")
type ProcessTemplate struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	Fields      []FieldDef `json:"fields,omitempty"` // Relasi ke definisi kolom
}

// FieldDef: Definisi kolom dinamis agar user bisa edit form sendiri
type FieldDef struct {
	ID             int    `json:"id" db:"id"`
	TemplateID     int    `json:"template_id" db:"template_id"`
	Key            string `json:"key" db:"field_key"`     // key untuk JSON
	Label          string `json:"label" db:"field_label"` // label UI
	Type           string `json:"type" db:"field_type"`   // number, text, date
	IsRequired     bool   `json:"required" db:"is_required"`
	ValidationRule string `json:"validation" db:"validation_rule"`
}

// ProcessInstance: Data nyata yang diinput operator
type ProcessInstance struct {
	ID           int64           `json:"id" db:"id"`
	TemplateID   int             `json:"template_id" db:"template_id"`
	WorkflowID   int             `json:"workflow_id" db:"workflow_id"`
	
	// DataPayload menampung JSON text dari database
	DataPayload  json.RawMessage `json:"data" db:"data_payload"`
	
	Status       string          `json:"status" db:"status"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

type Workflow struct {
	ID           int             `json:"id" db:"id"`
	Name         string          `json:"name" db:"name"`
	CanvasConfig json.RawMessage `json:"canvas_config" db:"canvas_config"` // Simpan posisi X,Y node
	IsActive     bool            `json:"is_active" db:"is_active"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}