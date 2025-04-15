package notification

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	// ErrEmptyTopic is returned when no topic is found
	ErrEmptyTopic = errors.New("no topic found")
)

// SignalRepository defines the interface for notification signal management
type SignalRepository interface {
	// Subscribe subscribes to a topic and returns a channel for notifications
	Subscribe(topic string) (<-chan *Notification, func(), error)

	// Publish publishes a notification to a topic
	Publish(topic string, notification *Notification) error
}

// Topic represents a notification topic
type Topic struct {
	Listeners []chan<- *Notification
	Mutex     *sync.Mutex
}

// signalRepository implements SignalRepository
type signalRepository struct {
	mutex     sync.Mutex
	topics    map[string]map[string]chan *Notification
	topicSize int
}

// NewSignalRepository creates a new signal repository
func NewSignalRepository(topicSize int) SignalRepository {
	return &signalRepository{
		mutex:     sync.Mutex{},
		topics:    make(map[string]map[string]chan *Notification),
		topicSize: topicSize,
	}
}

// Subscribe subscribes to a topic
func (r *signalRepository) Subscribe(topic string) (<-chan *Notification, func(), error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Create topic if it doesn't exist
	if _, exists := r.topics[topic]; !exists {
		r.topics[topic] = make(map[string]chan *Notification)
	}

	// Create a buffered channel for notifications
	ch := make(chan *Notification, r.topicSize)

	// Generate a unique subscriber ID
	subscriberID := uuid.New().String()

	// Add subscriber to the topic
	r.topics[topic][subscriberID] = ch

	// Create a cancel function
	cancel := func() {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		if topicMap, exists := r.topics[topic]; exists {
			delete(topicMap, subscriberID)

			// Clean up topic if no subscribers left
			if len(topicMap) == 0 {
				delete(r.topics, topic)
			}
		}

		close(ch)
	}

	return ch, cancel, nil
}

// Publish publishes a notification to a topic
func (r *signalRepository) Publish(topic string, notification *Notification) error {
	r.mutex.Lock()

	// Create topic if it doesn't exist
	if _, exists := r.topics[topic]; !exists {
		r.topics[topic] = make(map[string]chan *Notification)
		r.mutex.Unlock()
		return nil // No subscribers yet, so nothing to do
	}

	// Get a copy of subscribers to avoid deadlock while iterating
	subscribers := make([]chan *Notification, 0, len(r.topics[topic]))
	for _, ch := range r.topics[topic] {
		subscribers = append(subscribers, ch)
	}
	r.mutex.Unlock()

	// Log notification publishing
	subscriberCount := len(subscribers)
	if subscriberCount > 0 {
		logrus.WithFields(logrus.Fields{
			"notification_id": notification.ID,
			"topic":           topic,
			"subscribers":     subscriberCount,
		}).Debug("Publishing notification to subscribers")
	}

	// Publish to each subscriber in a non-blocking way
	for _, ch := range subscribers {
		// Use a goroutine to ensure we don't block if a channel is full
		go func(channel chan *Notification) {
			// Create a timeout context for sending
			select {
			case channel <- notification:
				// Successfully sent
			case <-time.After(100 * time.Millisecond):
				// Timed out - channel might be blocked
				logrus.WithFields(logrus.Fields{
					"notification_id": notification.ID,
					"topic":           topic,
				}).Warn("Failed to deliver notification to subscriber (channel full or blocked)")
			}
		}(ch)
	}

	return nil
}
