package broker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Common errors
var (
	ErrQueueNotFound     = errors.New("queue not found")
	ErrMessageProcessing = errors.New("message processing error")
	ErrQueueFull         = errors.New("queue is full")
	ErrSubscriptionError = errors.New("error creating subscription")
)

// Message represents a generic message in the message queue
type Message struct {
	ID          string                 `json:"id"`
	Topic       string                 `json:"topic"`
	Payload     []byte                 `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	PublishedAt time.Time              `json:"published_at"`
	Attributes  map[string]string      `json:"attributes,omitempty"`
}

// MessageHandler is a function that processes messages
type MessageHandler func(context.Context, *Message) error

// MessageBroker defines an interface for a message broker
type MessageBroker interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, payload []byte, attributes map[string]string) error

	// Subscribe subscribes to a topic with a handler function
	Subscribe(ctx context.Context, topic string, handler MessageHandler) (Subscription, error)

	// CreateTopic creates a new topic if it doesn't exist
	CreateTopic(ctx context.Context, topic string) error

	// DeleteTopic deletes a topic
	DeleteTopic(ctx context.Context, topic string) error

	// Close closes the message broker
	Close() error
}

// Subscription represents a subscription to a topic
type Subscription interface {
	// ID returns the subscription ID
	ID() string

	// Topic returns the topic name
	Topic() string

	// Unsubscribe unsubscribes from the topic
	Unsubscribe() error

	// IsClosed returns true if the subscription is closed
	IsClosed() bool
}

// InMemoryMessage represents a message in the in-memory broker
type InMemoryMessage struct {
	Message
	processed bool
	attempts  int
}

// InMemoryBroker is a simple in-memory implementation of MessageBroker
type InMemoryBroker struct {
	topics        map[string][]*InMemoryMessage
	subscriptions map[string]map[string]MessageHandler
	mu            sync.RWMutex
	logger        *logrus.Logger
	queueSize     int
	closed        bool
}

// subscription implements the Subscription interface
type subscription struct {
	id        string
	topic     string
	broker    *InMemoryBroker
	closed    bool
	closeChan chan struct{}
}

// NewInMemoryBroker creates a new in-memory message broker
func NewInMemoryBroker(logger *logrus.Logger, queueSize int) *InMemoryBroker {
	if queueSize <= 0 {
		queueSize = 1000 // Default queue size
	}

	broker := &InMemoryBroker{
		topics:        make(map[string][]*InMemoryMessage),
		subscriptions: make(map[string]map[string]MessageHandler),
		logger:        logger,
		queueSize:     queueSize,
	}

	return broker
}

// CreateTopic creates a new topic
func (b *InMemoryBroker) CreateTopic(ctx context.Context, topic string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return errors.New("broker is closed")
	}

	if _, exists := b.topics[topic]; !exists {
		b.topics[topic] = make([]*InMemoryMessage, 0)
		b.subscriptions[topic] = make(map[string]MessageHandler)
	}

	return nil
}

// DeleteTopic deletes a topic
func (b *InMemoryBroker) DeleteTopic(ctx context.Context, topic string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return errors.New("broker is closed")
	}

	if _, exists := b.topics[topic]; !exists {
		return ErrQueueNotFound
	}

	delete(b.topics, topic)
	delete(b.subscriptions, topic)

	return nil
}

// Publish publishes a message to a topic
func (b *InMemoryBroker) Publish(ctx context.Context, topic string, payload []byte, attributes map[string]string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return errors.New("broker is closed")
	}

	// Create topic if it doesn't exist
	if _, exists := b.topics[topic]; !exists {
		b.topics[topic] = make([]*InMemoryMessage, 0)
		b.subscriptions[topic] = make(map[string]MessageHandler)
	}

	// Check queue size limit
	if len(b.topics[topic]) >= b.queueSize {
		return ErrQueueFull
	}

	// Create and store the message
	msg := &InMemoryMessage{
		Message: Message{
			ID:          uuid.New().String(),
			Topic:       topic,
			Payload:     payload,
			PublishedAt: time.Now(),
			Attributes:  attributes,
		},
		processed: false,
		attempts:  0,
	}

	b.topics[topic] = append(b.topics[topic], msg)

	// Notify subscribers asynchronously
	if subs, ok := b.subscriptions[topic]; ok && len(subs) > 0 {
		for _, handler := range subs {
			go b.processMessage(ctx, handler, &msg.Message)
		}
	}

	return nil
}

// Subscribe subscribes to a topic
func (b *InMemoryBroker) Subscribe(ctx context.Context, topic string, handler MessageHandler) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, errors.New("broker is closed")
	}

	// Create topic if it doesn't exist
	if _, exists := b.topics[topic]; !exists {
		b.topics[topic] = make([]*InMemoryMessage, 0)
		b.subscriptions[topic] = make(map[string]MessageHandler)
	}

	// Create subscription
	subID := uuid.New().String()
	b.subscriptions[topic][subID] = handler

	sub := &subscription{
		id:        subID,
		topic:     topic,
		broker:    b,
		closeChan: make(chan struct{}),
	}

	return sub, nil
}

// processMessage processes a message with a handler
func (b *InMemoryBroker) processMessage(ctx context.Context, handler MessageHandler, msg *Message) {
	// Create a new background context for async processing to prevent cancellation issues
	processingCtx := context.Background()

	err := handler(processingCtx, msg)
	if err != nil {
		b.logger.WithError(err).WithField("message_id", msg.ID).Error("Error processing message")
	}
}

// Close closes the broker
func (b *InMemoryBroker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Clear all topics and subscriptions
	b.topics = nil
	b.subscriptions = nil

	return nil
}

// ID returns the subscription ID
func (s *subscription) ID() string {
	return s.id
}

// Topic returns the topic name
func (s *subscription) Topic() string {
	return s.topic
}

// IsClosed returns whether the subscription is closed
func (s *subscription) IsClosed() bool {
	return s.closed
}

// Unsubscribe unsubscribes from the topic
func (s *subscription) Unsubscribe() error {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()

	if s.closed {
		return nil
	}

	if subs, ok := s.broker.subscriptions[s.topic]; ok {
		delete(subs, s.id)
	}

	s.closed = true
	close(s.closeChan)

	return nil
}

// NotificationMessage defines the structure of a notification message
type NotificationMessage struct {
	NotificationID  string                 `json:"notification_id"`
	UserID          string                 `json:"user_id"`
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content"`
	Data            map[string]string      `json:"data,omitempty"`
	DeliveryMethods []string               `json:"delivery_methods,omitempty"`
	DeliveryConfig  map[string]interface{} `json:"delivery_config,omitempty"`
	Reference       string                 `json:"reference,omitempty"`
	ReferenceID     string                 `json:"reference_id,omitempty"`
}

// Helper method to create a notification message
func NewNotificationMessage(notificationID, userID, notificationType, title, content string, data map[string]string, methods []string) (*Message, error) {
	notifMsg := NotificationMessage{
		NotificationID:  notificationID,
		UserID:          userID,
		Type:            notificationType,
		Title:           title,
		Content:         content,
		Data:            data,
		DeliveryMethods: methods,
	}

	payload, err := json.Marshal(notifMsg)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:          uuid.New().String(),
		Topic:       "notifications",
		Payload:     payload,
		PublishedAt: time.Now(),
		Attributes: map[string]string{
			"type": notificationType,
			"user": userID,
		},
	}, nil
}
