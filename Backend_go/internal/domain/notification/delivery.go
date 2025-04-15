package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

// DeliveryServiceFactory creates delivery services for different methods
type DeliveryServiceFactory interface {
	// GetDeliveryService returns a delivery service for the given method
	GetDeliveryService(method DeliveryMethod) (DeliveryService, error)
}

// compositeDeliveryService implements DeliveryService with multiple delivery methods
type compositeDeliveryService struct {
	logger   *logrus.Logger
	factory  DeliveryServiceFactory
	inAppSvc DeliveryService
}

// NewCompositeDeliveryService creates a new composite delivery service
func NewCompositeDeliveryService(logger *logrus.Logger, factory DeliveryServiceFactory, inApp DeliveryService) DeliveryService {
	return &compositeDeliveryService{
		logger:   logger,
		factory:  factory,
		inAppSvc: inApp,
	}
}

// Deliver sends a notification through a specific channel
func (s *compositeDeliveryService) Deliver(ctx context.Context, notification *Notification, method DeliveryMethod) error {
	// Always use in-app for InApp method
	if method == InApp {
		return s.inAppSvc.Deliver(ctx, notification, method)
	}

	// Get appropriate delivery service from factory
	deliveryService, err := s.factory.GetDeliveryService(method)
	if err != nil {
		s.logger.WithError(err).WithField("method", method).
			Warn("Failed to get delivery service, skipping this method")
		return nil // Skip rather than fail the entire notification
	}

	// Skip nil delivery services (not yet implemented)
	if deliveryService == nil {
		s.logger.WithField("method", method).
			Info("Delivery service not implemented yet, skipping this method")
		return nil
	}

	// Deliver using the appropriate service
	return deliveryService.Deliver(ctx, notification, method)
}

// DeliverWithConfig sends a notification with specific configuration
func (s *compositeDeliveryService) DeliverWithConfig(ctx context.Context, notification *Notification, config map[string]interface{}) error {
	// Extract delivery methods from config
	methodsIface, ok := config["methods"]
	if !ok {
		// Default to in-app only
		return s.inAppSvc.Deliver(ctx, notification, InApp)
	}

	// Parse methods from interface
	methods, ok := methodsIface.([]DeliveryMethod)
	if !ok {
		// Try to parse as string array and convert
		methodsStr, ok := methodsIface.([]string)
		if !ok {
			return errors.New("invalid delivery methods format")
		}

		methods = make([]DeliveryMethod, len(methodsStr))
		for i, m := range methodsStr {
			methods[i] = DeliveryMethod(m)
		}
	}

	// Track errors but continue delivering to all methods
	var errs []error

	// Deliver to each method
	for _, method := range methods {
		// Extract method-specific config if available
		methodConfig, ok := config[string(method)]
		var methodConfigMap map[string]interface{}
		if ok {
			methodConfigMap, _ = methodConfig.(map[string]interface{})
		}

		// Get appropriate delivery service
		var svc DeliveryService
		if method == InApp {
			svc = s.inAppSvc
		} else {
			var err error
			svc, err = s.factory.GetDeliveryService(method)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}

		// If we have method-specific config, use it
		if methodConfigMap != nil {
			if err := svc.DeliverWithConfig(ctx, notification, methodConfigMap); err != nil {
				errs = append(errs, err)
			}
		} else {
			// Otherwise just use the method
			if err := svc.Deliver(ctx, notification, method); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// If we had any errors, return a combined error
	if len(errs) > 0 {
		errStr, _ := json.Marshal(errs)
		return fmt.Errorf("some delivery methods failed: %s", errStr)
	}

	return nil
}

// inAppDeliveryService implements DeliveryService for in-app notifications
type inAppDeliveryService struct {
	signalRepo SignalRepository
	logger     *logrus.Logger
}

// NewInAppDeliveryService creates a new in-app delivery service
func NewInAppDeliveryService(signalRepo SignalRepository, logger *logrus.Logger) DeliveryService {
	return &inAppDeliveryService{
		signalRepo: signalRepo,
		logger:     logger,
	}
}

// Deliver sends an in-app notification
func (s *inAppDeliveryService) Deliver(ctx context.Context, notification *Notification, method DeliveryMethod) error {
	// Publish notification to subscribers via WebSocket
	return s.signalRepo.Publish(notification.UserID.String(), notification)
}

// DeliverWithConfig sends an in-app notification with configuration
func (s *inAppDeliveryService) DeliverWithConfig(ctx context.Context, notification *Notification, config map[string]interface{}) error {
	// For in-app, we don't need special config
	return s.Deliver(ctx, notification, InApp)
}

// emailDeliveryService implements DeliveryService for email notifications
type emailDeliveryService struct {
	// Email service components would go here
	logger *logrus.Logger
}

// NewEmailDeliveryService creates a new email delivery service
func NewEmailDeliveryService(logger *logrus.Logger) DeliveryService {
	return &emailDeliveryService{
		logger: logger,
	}
}

// Deliver sends an email notification
func (s *emailDeliveryService) Deliver(ctx context.Context, notification *Notification, method DeliveryMethod) error {
	// In a real implementation, this would send an email
	s.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"title":           notification.Title,
	}).Info("Would send email notification")

	return nil
}

// DeliverWithConfig sends an email notification with configuration
func (s *emailDeliveryService) DeliverWithConfig(ctx context.Context, notification *Notification, config map[string]interface{}) error {
	// In a real implementation, this would use the config to customize the email
	s.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"title":           notification.Title,
		"config":          config,
	}).Info("Would send customized email notification")

	return nil
}

// Standard implementation of DeliveryServiceFactory
type standardDeliveryServiceFactory struct {
	emailSvc DeliveryService
	pushSvc  DeliveryService
	smsSvc   DeliveryService
	logger   *logrus.Logger
}

// NewStandardDeliveryServiceFactory creates a new standard delivery service factory
func NewStandardDeliveryServiceFactory(
	logger *logrus.Logger,
	emailSvc DeliveryService,
	pushSvc DeliveryService,
	smsSvc DeliveryService) DeliveryServiceFactory {

	return &standardDeliveryServiceFactory{
		emailSvc: emailSvc,
		pushSvc:  pushSvc,
		smsSvc:   smsSvc,
		logger:   logger,
	}
}

// GetDeliveryService returns a delivery service for the given method
func (f *standardDeliveryServiceFactory) GetDeliveryService(method DeliveryMethod) (DeliveryService, error) {
	switch method {
	case Email:
		if f.emailSvc == nil {
			return nil, fmt.Errorf("email delivery service not configured")
		}
		return f.emailSvc, nil
	case Push:
		if f.pushSvc == nil {
			return nil, fmt.Errorf("push delivery service not configured")
		}
		return f.pushSvc, nil
	case SMS:
		if f.smsSvc == nil {
			return nil, fmt.Errorf("SMS delivery service not configured")
		}
		return f.smsSvc, nil
	default:
		return nil, fmt.Errorf("unsupported delivery method: %s", method)
	}
}
