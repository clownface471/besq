package repository

import (
	"pt-besq-core/internal/database"
	"pt-besq-core/internal/entity"
)

// GetAllTemplates
func GetAllTemplates() ([]entity.ProcessTemplate, error) {
	var templates []entity.ProcessTemplate
	query := "SELECT id, name, description, created_at FROM process_templates"
	err := database.DB.Select(&templates, query)
	return templates, err
}

// GetFieldDefs mengambil aturan kolom (Ini adalah Single Source of Truth sekarang)
func GetFieldDefs(templateID int) ([]entity.FieldDef, error) {
	var fields []entity.FieldDef
	query := `
		SELECT id, template_id, field_key, field_label, field_type, is_required 
		FROM field_definitions 
		WHERE template_id = ?
	`
	err := database.DB.Select(&fields, query, templateID)
	return fields, err
}