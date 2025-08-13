package pubsub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message represents a pub/sub message
type Message struct {
	Id        string         `json:"id"`
	Type      string         `json:"type"`
	Topic     string         `json:"topic"`
	Data      map[string]any `json:"data"`
	Timestamp time.Time      `json:"timestamp"`
}

// MessageHandler processes incoming messages
type MessageHandler func(msg *Message) error

// PublisherConfig configures the publisher
type PublisherConfig struct {
	ServiceName  string
	IdentityPath string
}

// SubscriberConfig configures the subscriber
type SubscriberConfig struct {
	ServiceName    string
	IdentityPath   string
	ReconnectDelay time.Duration
	MaxReconnects  int
	MessageTimeout time.Duration
}

// DefaultSubscriberConfig returns sensible defaults
func DefaultSubscriberConfig() *SubscriberConfig {
	return &SubscriberConfig{
		ReconnectDelay: 5 * time.Second,
		MaxReconnects:  -1, // infinite
		MessageTimeout: 30 * time.Second,
	}
}

// Publisher publishes messages to subscribers
type Publisher interface {
	Publish(ctx context.Context, msg *Message) error
	Close() error
}

// Subscriber receives messages from publishers
type Subscriber interface {
	Subscribe(ctx context.Context, topics []string, handler MessageHandler) error
	Close() error
}

// NewMessage creates a new message
func NewMessage(msgType, topic string, data map[string]any) *Message {
	return &Message{
		Id:        uuid.New().String(),
		Type:      msgType,
		Topic:     topic,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Marshal converts message to JSON bytes
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalMessage converts JSON bytes to message
func UnmarshalMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
