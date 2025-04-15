package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// DomainNotifier provides a generic way for domains to create notifications
type DomainNotifier interface {
	// NotifyUser sends a notification to a specific user
	NotifyUser(ctx context.Context, userID uuid.UUID, notificationType Type, title, content string, data map[string]string, domain string, domainID uuid.UUID) error

	// NotifyUserWithDelivery sends a notification with specific delivery methods
	NotifyUserWithDelivery(ctx context.Context, userID uuid.UUID, notificationType Type, title, content string, data map[string]string, domain string, domainID uuid.UUID, methods []DeliveryMethod) error

	// NotifyUserWithConfig sends a notification with custom configuration
	NotifyUserWithConfig(ctx context.Context, userID uuid.UUID, notificationType Type, title, content string, data map[string]string, domain string, domainID uuid.UUID, config map[string]interface{}) error
}

// domainNotifierImpl implements DomainNotifier
type domainNotifierImpl struct {
	service  Service
	producer Producer
	logger   *logrus.Logger
}

// NewDomainNotifier creates a new domain notifier
func NewDomainNotifier(service Service, producer Producer, logger *logrus.Logger) DomainNotifier {
	return &domainNotifierImpl{
		service:  service,
		producer: producer,
		logger:   logger,
	}
}

// NotifyUser sends a notification to a specific user
func (n *domainNotifierImpl) NotifyUser(ctx context.Context, userID uuid.UUID, notificationType Type, title, content string, data map[string]string, domain string, domainID uuid.UUID) error {
	// Create notification directly via the service (non-brokered, synchronous)
	return n.service.CreateForUser(ctx, userID, notificationType, title, content, data, domain, domainID)
}

// NotifyUserWithDelivery sends a notification with specific delivery methods
func (n *domainNotifierImpl) NotifyUserWithDelivery(ctx context.Context, userID uuid.UUID, notificationType Type, title, content string, data map[string]string, domain string, domainID uuid.UUID, methods []DeliveryMethod) error {
	notification := &Notification{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        notificationType,
		Title:       title,
		Content:     content,
		Status:      Unread,
		Data:        data,
		Reference:   domain,
		ReferenceID: domainID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// If we have a producer, use the broker for asynchronous delivery
	if n.producer != nil {
		if err := n.producer.ProduceNotification(ctx, notification, methods); err != nil {
			n.logger.WithError(err).Error("Failed to produce notification")
			// Fall back to direct creation if broker fails
			return n.service.Create(ctx, notification)
		}
		return nil
	}

	// If no producer, use service directly
	if err := n.service.Create(ctx, notification); err != nil {
		return err
	}

	// If we have methods but no producer, try to deliver via service
	if len(methods) > 0 {
		return n.service.DeliverNotification(ctx, notification, methods)
	}

	return nil
}

// NotifyUserWithConfig sends a notification with custom configuration
func (n *domainNotifierImpl) NotifyUserWithConfig(ctx context.Context, userID uuid.UUID, notificationType Type, title, content string, data map[string]string, domain string, domainID uuid.UUID, config map[string]interface{}) error {
	notification := &Notification{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        notificationType,
		Title:       title,
		Content:     content,
		Status:      Unread,
		Data:        data,
		Reference:   domain,
		ReferenceID: domainID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// If we have a producer, use the broker for asynchronous delivery with config
	if n.producer != nil {
		if err := n.producer.ProduceNotificationWithConfig(ctx, notification, config); err != nil {
			n.logger.WithError(err).Error("Failed to produce notification with config")
			// Fall back to direct creation if broker fails
			return n.service.Create(ctx, notification)
		}
		return nil
	}

	// If no producer, use service directly
	return n.service.Create(ctx, notification)
}
