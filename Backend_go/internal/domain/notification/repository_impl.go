package notification

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// postgresRepository implements the Repository interface for PostgreSQL
type postgresRepository struct {
	db     *connection.Database
	logger *logrus.Logger
}

// NewRepository creates a new PostgreSQL notification repository
func NewRepository(db *connection.Database, logger *logrus.Logger) Repository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}

// withRecovery executes the given function with database error recovery
func (r *postgresRepository) withRecovery(ctx context.Context, operation string, fn func(tx *gorm.DB) error) error {
	// Get a database instance
	db := r.db.WithContext(ctx)

	// Execute the operation with recovery logic
	err := fn(db)

	// Handle specific database errors
	if err != nil {
		// Log details about the error
		r.logger.WithError(err).WithField("operation", operation).Error("Database operation failed")

		// Check if this is a connection error
		if isConnectionError(err) {
			r.logger.WithField("operation", operation).Warn("Database connection error, attempting reconnection")

			// Try to reconnect
			reconnectErr := r.db.Reconnect()
			if reconnectErr != nil {
				r.logger.WithError(reconnectErr).Error("Failed to reconnect to database")
				return err // Return the original error
			}

			// Get fresh DB instance and try the operation once more
			r.logger.WithField("operation", operation).Info("Reconnection successful, retrying operation")
			db = r.db.WithContext(ctx)
			retryErr := fn(db)
			if retryErr != nil {
				r.logger.WithError(retryErr).Error("Operation failed after reconnection")
				return retryErr
			}

			return nil // Success on retry
		}

		return err
	}

	return nil
}

// Check if an error is related to connection problems
func isConnectionError(err error) bool {
	errMsg := err.Error()
	return contains(errMsg, "connection refused") || contains(errMsg, "bad connection") || contains(errMsg, "connection reset by peer") || contains(errMsg, "broken pipe") || contains(errMsg, "connection closed") || contains(errMsg, "driver: bad connection") || contains(errMsg, "hostname resolving error") || contains(errMsg, "operation was canceled")
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Create creates a new notification
func (r *postgresRepository) Create(ctx context.Context, notification *Notification) error {
	return r.withRecovery(ctx, "Create", func(tx *gorm.DB) error {
		return tx.Create(notification).Error
	})
}

// GetByID retrieves a notification by its ID
func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Notification, error) {
	var notification Notification

	err := r.withRecovery(ctx, "GetByID", func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).First(&notification)
		return result.Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ErrNotFoundType{Message: "notification not found"}
		}
		return nil, err
	}

	return &notification, nil
}

// GetByUserID retrieves all notifications for a user
func (r *postgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error) {
	var notifications []*Notification

	err := r.withRecovery(ctx, "GetByUserID", func(tx *gorm.DB) error {
		// Build query
		query := tx.Model(&Notification{}).
			Where("user_id = ?", userID).
			Order("status ASC, created_at DESC") // Sort unread first, then by date

		if limit > 0 {
			query = query.Limit(limit)
		}

		if offset > 0 {
			query = query.Offset(offset)
		}

		return query.Find(&notifications).Error
	})

	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// GetUnreadByUserID retrieves unread notifications for a user
func (r *postgresRepository) GetUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error) {
	var notifications []*Notification

	err := r.withRecovery(ctx, "GetUnreadByUserID", func(tx *gorm.DB) error {
		query := tx.Where("user_id = ? AND status = ?", userID, Unread).
			Order("created_at DESC")

		if limit > 0 {
			query = query.Limit(limit)
		}

		if offset > 0 {
			query = query.Offset(offset)
		}

		return query.Find(&notifications).Error
	})

	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// UpdateStatus updates the status of a notification
func (r *postgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error {
	return r.withRecovery(ctx, "UpdateStatus", func(tx *gorm.DB) error {
		result := tx.Model(&Notification{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":     status,
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrNotFound
		}

		return nil
	})
}

// MarkAsRead marks a notification as read
func (r *postgresRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.withRecovery(ctx, "MarkAsRead", func(tx *gorm.DB) error {
		result := tx.Model(&Notification{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":     Read,
				"read_at":    now,
				"updated_at": now,
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrNotFound
		}

		return nil
	})
}

// MarkAllAsRead marks all notifications as read for a user
func (r *postgresRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	return r.withRecovery(ctx, "MarkAllAsRead", func(tx *gorm.DB) error {
		return tx.Model(&Notification{}).
			Where("user_id = ? AND status = ?", userID, Unread).
			Updates(map[string]interface{}{
				"status":     Read,
				"read_at":    now,
				"updated_at": now,
			}).Error
	})
}

// Delete removes a notification
func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.withRecovery(ctx, "Delete", func(tx *gorm.DB) error {
		result := tx.Delete(&Notification{}, "id = ?", id)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrNotFound
		}

		return nil
	})
}

// CountUnread counts unread notifications for a user
func (r *postgresRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	err := r.withRecovery(ctx, "CountUnread", func(tx *gorm.DB) error {
		return tx.Model(&Notification{}).
			Where("user_id = ? AND status = ?", userID, Unread).
			Count(&count).Error
	})

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// DeleteExpired removes all expired notifications
func (r *postgresRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()

	return r.withRecovery(ctx, "DeleteExpired", func(tx *gorm.DB) error {
		return tx.Where("expires_at IS NOT NULL AND expires_at < ?", now).
			Delete(&Notification{}).Error
	})
}
