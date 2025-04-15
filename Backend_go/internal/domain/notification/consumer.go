package notification

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/broker"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Consumer defines an interface for consuming notifications
type Consumer interface {
	// Start starts the consumer
	Start(ctx context.Context) error

	// Stop stops the consumer
	Stop() error

	// IsRunning returns true if the consumer is running
	IsRunning() bool
}

// brokerConsumer implements the Consumer interface
type brokerConsumer struct {
	messageBroker   broker.MessageBroker
	deliveryService DeliveryService
	repository      Repository
	logger          *logrus.Logger
	topicName       string
	subscription    broker.Subscription
	isRunning       bool
}

// NewBrokerConsumer creates a new notification consumer
func NewBrokerConsumer(
	messageBroker broker.MessageBroker,
	deliveryService DeliveryService,
	repository Repository,
	logger *logrus.Logger) Consumer {

	return &brokerConsumer{
		messageBroker:   messageBroker,
		deliveryService: deliveryService,
		repository:      repository,
		logger:          logger,
		topicName:       "notifications",
		isRunning:       false,
	}
}

// Start starts the consumer
func (c *brokerConsumer) Start(ctx context.Context) error {
	if c.isRunning {
		return errors.New("consumer already running")
	}

	// Ensure topic exists
	if err := c.messageBroker.CreateTopic(ctx, c.topicName); err != nil {
		c.logger.WithError(err).Error("Failed to create notifications topic")
		return err
	}

	// Create subscription
	sub, err := c.messageBroker.Subscribe(ctx, c.topicName, c.handleMessage)
	if err != nil {
		c.logger.WithError(err).Error("Failed to subscribe to notifications topic")
		return err
	}

	c.subscription = sub
	c.isRunning = true

	c.logger.WithFields(logrus.Fields{
		"topic":        c.topicName,
		"subscription": sub.ID(),
	}).Info("Notification consumer started")

	return nil
}

// Stop stops the consumer
func (c *brokerConsumer) Stop() error {
	if !c.isRunning {
		return nil
	}

	if c.subscription != nil {
		if err := c.subscription.Unsubscribe(); err != nil {
			c.logger.WithError(err).Error("Failed to unsubscribe from notifications topic")
			return err
		}
	}

	c.isRunning = false
	c.subscription = nil

	c.logger.Info("Notification consumer stopped")

	return nil
}

// IsRunning returns true if the consumer is running
func (c *brokerConsumer) IsRunning() bool {
	return c.isRunning
}

// handleMessage processes a notification message
func (c *brokerConsumer) handleMessage(ctx context.Context, message *broker.Message) error {
	c.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"topic":      message.Topic,
	}).Debug("Received notification message")

	// Create a new background context so it won't be canceled when the request completes
	processingCtx := context.Background()

	// Parse message payload
	var notifMsg broker.NotificationMessage
	if err := json.Unmarshal(message.Payload, &notifMsg); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal notification message")
		return err
	}

	// Parse UUIDs
	notificationID, err := uuid.Parse(notifMsg.NotificationID)
	if err != nil {
		c.logger.WithError(err).Error("Invalid notification ID in message")
		return err
	}

	userID, err := uuid.Parse(notifMsg.UserID)
	if err != nil {
		c.logger.WithError(err).Error("Invalid user ID in message")
		return err
	}

	// Check if notification exists in repository
	notification, err := c.repository.GetByID(processingCtx, notificationID)
	if err != nil {
		// If not found, create it
		var notFoundErr *ErrNotFoundType
		if errors.As(err, &notFoundErr) {
			notification = &Notification{
				ID:        notificationID,
				UserID:    userID,
				Type:      Type(notifMsg.Type),
				Title:     notifMsg.Title,
				Content:   notifMsg.Content,
				Status:    Unread,
				Data:      notifMsg.Data,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Add reference fields if available
			if notifMsg.Reference != "" {
				notification.Reference = notifMsg.Reference
			}
			if notifMsg.ReferenceID != "" {
				refID, parseErr := uuid.Parse(notifMsg.ReferenceID)
				if parseErr == nil {
					notification.ReferenceID = refID
				}
			}

			if err := c.repository.Create(processingCtx, notification); err != nil {
				c.logger.WithError(err).Error("Failed to create notification")
				return err
			}
		} else {
			c.logger.WithError(err).Error("Failed to check if notification exists")
			return err
		}
	}

	// Process delivery methods
	var methods []DeliveryMethod
	if len(notifMsg.DeliveryMethods) > 0 {
		methods = make([]DeliveryMethod, len(notifMsg.DeliveryMethods))
		for i, method := range notifMsg.DeliveryMethods {
			methods[i] = DeliveryMethod(method)
		}
	} else {
		// Default to in-app only
		methods = []DeliveryMethod{InApp}
	}

	// Track delivery errors but don't fail the whole message
	var deliveryErrors []error

	// Process with delivery service
	if notifMsg.DeliveryConfig != nil {
		err = c.deliveryService.DeliverWithConfig(processingCtx, notification, notifMsg.DeliveryConfig)
		if err != nil {
			c.logger.WithError(err).Error("Failed to deliver notification with custom config")
			deliveryErrors = append(deliveryErrors, err)
			// Continue processing - we already stored the notification
		}
	} else {
		// Deliver to each method
		for _, method := range methods {
			if err := c.deliveryService.Deliver(processingCtx, notification, method); err != nil {
				c.logger.WithError(err).WithField("method", method).
					Error("Failed to deliver notification")
				deliveryErrors = append(deliveryErrors, err)
				// Continue with other methods
			}
		}
	}

	// Log success if no errors, or summarize errors if some methods failed
	if len(deliveryErrors) == 0 {
		c.logger.WithFields(logrus.Fields{
			"notification_id": notification.ID,
			"user_id":         notification.UserID,
			"methods":         methods,
		}).Debug("Notification delivered successfully")
	} else {
		c.logger.WithFields(logrus.Fields{
			"notification_id": notification.ID,
			"user_id":         notification.UserID,
			"methods":         methods,
			"error_count":     len(deliveryErrors),
		}).Warn("Notification delivered with some errors")
	}

	// We always return nil to avoid reprocessing the message
	// since we've already stored it in the database
	return nil
}
