package handler

import (
	"net/http"
	"pt-besq-core/internal/entity"
	"pt-besq-core/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	Repo *repository.WorkflowRepository
}

func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{
		Repo: repository.NewWorkflowRepository(),
	}
}

func (h *WorkflowHandler) GetList(c *gin.Context) {
	list, err := h.Repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *WorkflowHandler) Create(c *gin.Context) {
	var input entity.Workflow
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.Repo.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Workflow created", "id": id})
}

func (h *WorkflowHandler) UpdateLayout(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// Kita baca raw body sebagai string JSON
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}
	
	err = h.Repo.UpdateLayout(id, string(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Layout updated"})
}