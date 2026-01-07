package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pt-besq-core/internal/repository"
	"pt-besq-core/internal/websocket"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type InstanceHandler struct {
	Repo *repository.InstanceRepository
	Hub  *websocket.Hub
}

func NewInstanceHandler(hub *websocket.Hub) *InstanceHandler {
	return &InstanceHandler{
		Repo: repository.NewInstanceRepository(),
		Hub:  hub,
	}
}

// CreateInstance (Input Data)
func (h *InstanceHandler) CreateInstance(c *gin.Context) {
	var req struct {
		TemplateID int                    `json:"template_id"`
		WorkflowID int                    `json:"workflow_id"`
		Data       map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fields, err := repository.GetFieldDefs(req.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template tidak valid"})
		return
	}

	// Memanggil Repo ValidateInput
	if err := h.Repo.ValidateInput(req.Data, fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonBytes, _ := json.Marshal(req.Data)
	
	// Memanggil Repo SaveInstance
	id, err := h.Repo.SaveInstance(req.WorkflowID, req.TemplateID, jsonBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data"})
		return
	}

	// Kirim pesan ke WebSocket (Struct Message sudah didefinisikan di Hub)
	msg := websocket.Message{
		Event:      "new_data",
		InstanceID: int(id),
		WorkflowID: req.WorkflowID,
		TemplateID: req.TemplateID,
		Status:     "draft",
		Timestamp:  time.Now(),
	}
	h.Hub.Broadcast <- msg

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Data valid, saved & broadcasted",
		"id":        id,
		"timestamp": time.Now(),
	})
}

// GetList (History & Pagination)
func (h *InstanceHandler) GetList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	templateID, _ := strconv.Atoi(c.Query("template_id"))
	dateStr := c.Query("date")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 { page = 1 }
	offset := (page - 1) * limit

	logs, err := h.Repo.GetHistory(limit, offset, templateID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil data: " + err.Error()})
		return
	}

	total, _ := h.Repo.CountHistory(templateID, dateStr)

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_data":   total,
			"total_pages":  (total + limit - 1) / limit,
		},
	})
}

// ExportExcel (Download .xlsx)
func (h *InstanceHandler) ExportExcel(c *gin.Context) {
	templateID, _ := strconv.Atoi(c.Query("template_id"))
	dateStr := c.Query("date")

	// Limit besar untuk ambil semua data
	logs, err := h.Repo.GetHistory(10000, 0, templateID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil data: " + err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "Data Produksi"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") 

	headers := []string{"ID", "Waktu", "Proses", "Workflow", "Status", "Data JSON"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	for i, log := range logs {
		rowNum := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), log.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), log.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), log.TemplateName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), log.WorkflowName)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowNum), log.Status)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), string(log.DataPayload))
	}

	filename := fmt.Sprintf("Laporan_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	
	// FIX: WriteTo mengembalikan 2 value (int64, error). Kita pakai underscore (_)
	if _, err := f.WriteTo(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate file"})
	}
}