package dto

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/notification"
	"github.com/google/uuid"
)

// NotificationFilter represents request filtering parameters for notifications
type NotificationFilter struct {
	Status    string     `form:"status"`
	Type      string     `form:"type"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	Page      int        `form:"page,default=0"`
	PageSize  int        `form:"page_size,default=10"`
}

// NotificationDTO represents a notification data transfer object
type NotificationDTO struct {
	ID          uuid.UUID         `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Status      string            `json:"status"`
	Data        map[string]string `json:"data,omitempty"`
	Reference   string            `json:"reference,omitempty"`
	ReferenceID uuid.UUID         `json:"reference_id,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ReadAt      *time.Time        `json:"read_at,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID      uuid.UUID         `json:"user_id" binding:"required"`
	Type        string            `json:"type" binding:"required"`
	Title       string            `json:"title" binding:"required"`
	Content     string            `json:"content" binding:"required"`
	Data        map[string]string `json:"data,omitempty"`
	Reference   string            `json:"reference,omitempty"`
	ReferenceID uuid.UUID         `json:"reference_id,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// NotificationUpdateRequest represents a request to update a notification
type NotificationUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=UNREAD READ ARCHIVED"`
}

// NotificationListResponse represents a paginated response of notifications
type NotificationListResponse struct {
	Items       []NotificationDTO `json:"items"`
	TotalCount  int               `json:"total_count"`
	UnreadCount int               `json:"unread_count"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
}

// NotificationCountResponse represents the count of notifications
type NotificationCountResponse struct {
	UnreadCount int `json:"unread_count"`
	TotalCount  int `json:"total_count"`
}

// ToDTO converts a domain notification model to a DTO
func ToDTO(n *notification.Notification) NotificationDTO {
	return NotificationDTO{
		ID:          n.ID,
		UserID:      n.UserID,
		Type:        string(n.Type),
		Title:       n.Title,
		Content:     n.Content,
		Status:      string(n.Status),
		Data:        n.Data,
		Reference:   n.Reference,
		ReferenceID: n.ReferenceID,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
		ReadAt:      n.ReadAt,
		ExpiresAt:   n.ExpiresAt,
	}
}

// ToDTOs converts a slice of domain notification models to DTOs
func ToDTOs(notifications []*notification.Notification) []NotificationDTO {
	dtos := make([]NotificationDTO, len(notifications))
	for i, n := range notifications {
		dtos[i] = ToDTO(n)
	}
	return dtos
}

// ToModel converts a DTO to a domain notification model
func (dto *CreateNotificationRequest) ToModel() *notification.Notification {
	return &notification.Notification{
		ID:          uuid.New(),
		UserID:      dto.UserID,
		Type:        notification.Type(dto.Type),
		Title:       dto.Title,
		Content:     dto.Content,
		Status:      notification.Unread,
		Data:        dto.Data,
		Reference:   dto.Reference,
		ReferenceID: dto.ReferenceID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiresAt:   dto.ExpiresAt,
	}
}
