package handler

import (
	"net/http"
	"pt-besq-core/internal/repository"
	"pt-besq-core/internal/websocket"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	Repo *repository.NotificationRepository
	Hub  *websocket.Hub
}

func NewNotificationHandler(hub *websocket.Hub) *NotificationHandler {
	return &NotificationHandler{
		Repo: repository.NewNotificationRepository(),
		Hub:  hub,
	}
}

// GetNotifications retrieves user notifications with pagination
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// Try to get user_id from context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert to int (handle both int and float64)
	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	unreadOnlyStr := c.DefaultQuery("unread_only", "false")
	
	limit, _ := strconv.Atoi(limitStr)
	unreadOnly := unreadOnlyStr == "true"

	notifications, err := h.Repo.GetUserNotifications(userID, limit, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	// Get unread count
	unreadCount, _ := h.Repo.GetUnreadCount(userID)

	c.JSON(http.StatusOK, gin.H{
		"data":         notifications,
		"unread_count": unreadCount,
	})
}

// MarkAsRead marks a single notification as read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	notifIDStr := c.Param("id")
	notifID, err := strconv.Atoi(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.Repo.MarkAsRead(notifID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead marks all user notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	err := h.Repo.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// SendNotification creates and broadcasts a new notification
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req struct {
		UserID            int    `json:"user_id" binding:"required"`
		Type              string `json:"type" binding:"required"`
		Title             string `json:"title" binding:"required"`
		Message           string `json:"message" binding:"required"`
		RelatedEntityType string `json:"related_entity_type"`
		RelatedEntityID   *int64 `json:"related_entity_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notif := repository.Notification{
		UserID:            req.UserID,
		Type:              req.Type,
		Title:             req.Title,
		Message:           req.Message,
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
	}

	id, err := h.Repo.Create(notif)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	// Broadcast via WebSocket
	wsMsg := websocket.Message{
		Event: "new_notification",
		Data: map[string]interface{}{
			"id":      id,
			"user_id": req.UserID,
			"type":    req.Type,
			"title":   req.Title,
			"message": req.Message,
		},
	}
	h.Hub.BroadcastToUser <- websocket.UserMessage{
		UserID:  req.UserID,
		Message: wsMsg,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification sent",
		"id":      id,
	})
}

// BroadcastToRole sends notification to all users with specific role
func (h *NotificationHandler) BroadcastToRole(c *gin.Context) {
	var req struct {
		Role              string `json:"role" binding:"required"`
		Type              string `json:"type" binding:"required"`
		Title             string `json:"title" binding:"required"`
		Message           string `json:"message" binding:"required"`
		RelatedEntityType string `json:"related_entity_type"`
		RelatedEntityID   *int64 `json:"related_entity_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notif := repository.Notification{
		Type:              req.Type,
		Title:             req.Title,
		Message:           req.Message,
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
	}

	err := h.Repo.BroadcastToRole(req.Role, notif)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to broadcast notification"})
		return
	}

	// Broadcast via WebSocket to all users with this role
	wsMsg := websocket.Message{
		Event: "new_notification",
		Data: map[string]interface{}{
			"role":    req.Role,
			"type":    req.Type,
			"title":   req.Title,
			"message": req.Message,
		},
	}
	h.Hub.BroadcastToRole <- websocket.RoleMessage{
		Role:    req.Role,
		Message: wsMsg,
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification broadcasted to role: " + req.Role})
}