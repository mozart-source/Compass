package main

import (
	"context"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/notification"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/broker"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

// NotificationSystem holds all notification-related components
type NotificationSystem struct {
	Service          notification.Service
	Producer         notification.Producer
	Consumer         notification.Consumer
	SignalRepository notification.SignalRepository
	MessageBroker    broker.MessageBroker
	Logger           *logrus.Logger
	CancelFunc       context.CancelFunc
	DomainNotifier   notification.DomainNotifier
}

// SetupNotificationSystem initializes and configures all notification components
func SetupNotificationSystem(
	db *connection.Database,
	appLogger *logger.Logger,
	isDevelopment bool,
) (*NotificationSystem, error) {
	// Initialize logger
	notifLogger := logrus.New()
	notifLogger.SetFormatter(&logrus.JSONFormatter{})
	if isDevelopment {
		notifLogger.SetLevel(logrus.DebugLevel)
	} else {
		notifLogger.SetLevel(logrus.InfoLevel)
	}

	// Initialize repositories
	repo := notification.NewRepository(db, notifLogger)
	signalRepo := notification.NewSignalRepository(100) // Buffer size of 100

	// Initialize message broker
	msgBroker := broker.NewInMemoryBroker(notifLogger, 1000) // Buffer 1000 messages

	// Initialize delivery services
	inAppDelivery := notification.NewInAppDeliveryService(signalRepo, notifLogger)
	emailDelivery := notification.NewEmailDeliveryService(notifLogger)

	// Create delivery factory
	deliveryFactory := notification.NewStandardDeliveryServiceFactory(
		notifLogger,
		emailDelivery,
		nil, // Push service (implement later)
		nil, // SMS service (implement later)
	)

	// Initialize composite delivery service
	compositeDelivery := notification.NewCompositeDeliveryService(
		notifLogger,
		deliveryFactory,
		inAppDelivery,
	)

	// Initialize delivery services map
	deliveryServices := map[notification.DeliveryMethod]notification.DeliveryService{
		notification.InApp: inAppDelivery,
		notification.Email: emailDelivery,
	}

	// Initialize notification service
	service := notification.NewService(notification.ServiceConfig{
		Repository:       repo,
		Logger:           notifLogger,
		SignalRepo:       signalRepo,
		DeliveryServices: deliveryServices,
	})

	// Initialize producer and consumer
	producer := notification.NewBrokerProducer(msgBroker, notifLogger)
	consumer := notification.NewBrokerConsumer(
		msgBroker,
		compositeDelivery,
		repo,
		notifLogger,
	)

	// Initialize domain notifier for use by different domains
	domainNotifier := notification.NewDomainNotifier(service, producer, notifLogger)

	// Start consumer in background
	consumerCtx, cancelFunc := context.WithCancel(context.Background())
	if err := consumer.Start(consumerCtx); err != nil {
		cancelFunc()
		appLogger.Error("Failed to start notification consumer", zap.Error(err))
		return nil, err
	}

	appLogger.Info("Notification system started successfully")

	return &NotificationSystem{
		Service:          service,
		Producer:         producer,
		Consumer:         consumer,
		SignalRepository: signalRepo,
		MessageBroker:    msgBroker,
		Logger:           notifLogger,
		CancelFunc:       cancelFunc,
		DomainNotifier:   domainNotifier,
	}, nil
}

// Shutdown gracefully stops all notification components
func (ns *NotificationSystem) Shutdown() error {
	if ns.Consumer != nil && ns.Consumer.IsRunning() {
		if err := ns.Consumer.Stop(); err != nil {
			ns.Logger.WithError(err).Error("Error shutting down notification consumer")
			return err
		}
	}

	if ns.CancelFunc != nil {
		ns.CancelFunc()
	}

	if ns.MessageBroker != nil {
		if err := ns.MessageBroker.Close(); err != nil {
			ns.Logger.WithError(err).Error("Error closing message broker")
			return err
		}
	}

	ns.Logger.Info("Notification system shut down successfully")
	return nil
}
