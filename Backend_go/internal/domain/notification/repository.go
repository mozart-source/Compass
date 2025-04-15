package notification

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for notification data access
type Repository interface {
	Create(ctx context.Context, notification *Notification) error

	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)

	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error)

	GetUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error)

	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error

	MarkAsRead(ctx context.Context, id uuid.UUID) error

	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error

	Delete(ctx context.Context, id uuid.UUID) error

	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)

	DeleteExpired(ctx context.Context) error
}
