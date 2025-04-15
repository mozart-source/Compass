package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/broker"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Producer defines an interface for producing notifications
type Producer interface {
	// ProduceNotification enqueues a notification for delivery
	ProduceNotification(ctx context.Context, notification *Notification, methods []DeliveryMethod) error

	// ProduceNotificationWithConfig enqueues a notification with custom config
	ProduceNotificationWithConfig(ctx context.Context, notification *Notification, config map[string]interface{}) error
}

// brokerProducer implements the Producer interface using a message broker
type brokerProducer struct {
	messageBroker broker.MessageBroker
	logger        *logrus.Logger
	topicName     string
}

// NewBrokerProducer creates a new notification producer using a message broker
func NewBrokerProducer(messageBroker broker.MessageBroker, logger *logrus.Logger) Producer {
	p := &brokerProducer{
		messageBroker: messageBroker,
		logger:        logger,
		topicName:     "notifications",
	}

	// Ensure the notifications topic exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := messageBroker.CreateTopic(ctx, p.topicName); err != nil {
		logger.WithError(err).Error("Failed to create notifications topic")
	}

	return p
}

// ProduceNotification enqueues a notification for delivery
func (p *brokerProducer) ProduceNotification(ctx context.Context, notification *Notification, methods []DeliveryMethod) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Convert delivery methods to strings
	methodsStr := make([]string, len(methods))
	for i, method := range methods {
		methodsStr[i] = string(method)
	}

	// Create broker message
	message, err := createNotificationMessage(notification, methodsStr, nil)
	if err != nil {
		p.logger.WithError(err).Error("Failed to create notification message")
		return err
	}

	// Publish to broker
	err = p.messageBroker.Publish(ctx, p.topicName, message.Payload, message.Attributes)
	if err != nil {
		p.logger.WithError(err).Error("Failed to publish notification message")
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"message_id":      message.ID,
	}).Debug("Notification message published")

	return nil
}

// ProduceNotificationWithConfig enqueues a notification with custom config
func (p *brokerProducer) ProduceNotificationWithConfig(ctx context.Context, notification *Notification, config map[string]interface{}) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Extract delivery methods from config
	methodsIface, ok := config["methods"]
	var methodsStr []string

	if ok {
		// Try to parse as DeliveryMethod array
		methods, ok := methodsIface.([]DeliveryMethod)
		if ok {
			methodsStr = make([]string, len(methods))
			for i, method := range methods {
				methodsStr[i] = string(method)
			}
		} else {
			// Try to parse as string array
			methodsStr, _ = methodsIface.([]string)
		}
	}

	// Default to in-app only if no methods specified
	if methodsStr == nil {
		methodsStr = []string{string(InApp)}
	}

	// Create broker message
	message, err := createNotificationMessage(notification, methodsStr, config)
	if err != nil {
		p.logger.WithError(err).Error("Failed to create notification message")
		return err
	}

	// Publish to broker
	err = p.messageBroker.Publish(ctx, p.topicName, message.Payload, message.Attributes)
	if err != nil {
		p.logger.WithError(err).Error("Failed to publish notification message")
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"message_id":      message.ID,
	}).Debug("Notification message with config published")

	return nil
}

// Helper to create a notification message
func createNotificationMessage(notification *Notification, methods []string, config map[string]interface{}) (*broker.Message, error) {
	// Create notification message
	notifMsg := broker.NotificationMessage{
		NotificationID:  notification.ID.String(),
		UserID:          notification.UserID.String(),
		Type:            string(notification.Type),
		Title:           notification.Title,
		Content:         notification.Content,
		Data:            notification.Data,
		DeliveryMethods: methods,
	}

	// Add reference fields if available
	if notification.Reference != "" {
		notifMsg.Reference = notification.Reference
	}
	if notification.ReferenceID != uuid.Nil {
		notifMsg.ReferenceID = notification.ReferenceID.String()
	}

	// Add delivery config if provided
	if config != nil {
		notifMsg.DeliveryConfig = config
	}

	// Marshal to JSON
	payload, err := json.Marshal(notifMsg)
	if err != nil {
		return nil, err
	}

	// Create message
	return &broker.Message{
		ID:          uuid.New().String(),
		Topic:       "notifications",
		Payload:     payload,
		PublishedAt: time.Now(),
		Attributes: map[string]string{
			"type": string(notification.Type),
			"user": notification.UserID.String(),
		},
	}, nil
}
