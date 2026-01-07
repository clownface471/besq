package handler

import (
	"net/http"
	"pt-besq-core/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct{}

func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{}
}

// GetAll mengambil semua jenis proses (Mixing, Oven, dll)
func (h *TemplateHandler) GetAll(c *gin.Context) {
	data, err := repository.GetAllTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetFields mengambil "Resep Form" (kolom apa saja yang dibutuhkan)
// Endpoint: GET /api/templates/:id/fields
func (h *TemplateHandler) GetFields(c *gin.Context) {
	// 1. Ambil ID dari URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Template harus angka"})
		return
	}

	// 2. Ambil definisi kolom dari Repository
	// (Fungsi GetFieldDefs sudah kita buat saat bikin Validasi tadi)
	fields, err := repository.GetFieldDefs(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil definisi kolom: " + err.Error()})
		return
	}

	// 3. Kembalikan JSON Schema ke Frontend
	c.JSON(http.StatusOK, gin.H{
		"template_id": id,
		"fields":      fields,
	})
}