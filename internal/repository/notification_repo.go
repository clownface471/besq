package repository

import (
	"pt-besq-core/internal/database"
	"time"
)

// Notification represents a system notification
type Notification struct {
	ID                int       `db:"id" json:"id"`
	UserID            int       `db:"user_id" json:"user_id"`
	Type              string    `db:"type" json:"type"` // info, warning, error, success
	Title             string    `db:"title" json:"title"`
	Message           string    `db:"message" json:"message"`
	RelatedEntityType string    `db:"related_entity_type" json:"related_entity_type,omitempty"`
	RelatedEntityID   *int64    `db:"related_entity_id" json:"related_entity_id,omitempty"`
	IsRead            bool      `db:"is_read" json:"is_read"`
	ReadAt            *time.Time `db:"read_at" json:"read_at,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

type NotificationRepository struct{}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{}
}

// Create creates a new notification
func (r *NotificationRepository) Create(notif Notification) (int64, error) {
	query := `
		INSERT INTO notifications (user_id, type, title, message, related_entity_type, related_entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	result, err := database.DB.Exec(query,
		notif.UserID, notif.Type, notif.Title, notif.Message,
		notif.RelatedEntityType, notif.RelatedEntityID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetUserNotifications retrieves notifications for a specific user
func (r *NotificationRepository) GetUserNotifications(userID int, limit int, unreadOnly bool) ([]Notification, error) {
	var notifications []Notification
	query := `
		SELECT id, user_id, type, title, message, related_entity_type, 
		       related_entity_id, is_read, read_at, created_at
		FROM notifications
		WHERE user_id = ?
	`
	args := []interface{}{userID}
	
	if unreadOnly {
		query += " AND is_read = 0"
	}
	
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	
	err := database.DB.Select(&notifications, query, args...)
	return notifications, err
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(notificationID int, userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = 1, read_at = NOW() 
		WHERE id = ? AND user_id = ?
	`
	_, err := database.DB.Exec(query, notificationID, userID)
	return err
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = 1, read_at = NOW() 
		WHERE user_id = ? AND is_read = 0
	`
	_, err := database.DB.Exec(query, userID)
	return err
}

// GetUnreadCount returns the count of unread notifications
func (r *NotificationRepository) GetUnreadCount(userID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0"
	err := database.DB.Get(&count, query, userID)
	return count, err
}

// DeleteOld deletes notifications older than specified days
func (r *NotificationRepository) DeleteOld(days int) error {
	query := "DELETE FROM notifications WHERE created_at < DATE_SUB(NOW(), INTERVAL ? DAY)"
	_, err := database.DB.Exec(query, days)
	return err
}

// BroadcastToRole sends notification to all users with specific role
func (r *NotificationRepository) BroadcastToRole(role string, notif Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, title, message, related_entity_type, related_entity_id, created_at)
		SELECT id, ?, ?, ?, ?, ?, NOW()
		FROM users
		WHERE role = ? AND is_active = 1
	`
	_, err := database.DB.Exec(query,
		notif.Type, notif.Title, notif.Message,
		notif.RelatedEntityType, notif.RelatedEntityID, role,
	)
	return err
}