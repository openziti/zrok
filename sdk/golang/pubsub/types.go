package pubsub

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

// Message represents a pub/sub message
type Message struct {
	Id         string         `json:"id"`
	Type       string         `json:"type"`
	Topic      string         `json:"topic"`
	Data       map[string]any `json:"data"`
	Timestamp  time.Time      `json:"timestamp"`
	OriginNode string         `json:"origin_node,omitempty"` // for mesh deduplication
	HopCount   int            `json:"hop_count,omitempty"`   // prevent infinite loops
}

// MessageHandler processes incoming messages
type MessageHandler func(msg *Message) error

// PublisherConfig configures the publisher
type PublisherConfig struct {
	ServiceName  string
	IdentityPath string
}

// MeshConfig configures mesh publisher/subscriber behavior
type MeshConfig struct {
	ServiceName       string
	IdentityPath      string
	NodeID           string
	PeerRefreshDelay time.Duration
	MaxHops          int
	MessageCacheSize int
	MessageCacheTTL  time.Duration
}

// DefaultMeshConfig returns sensible defaults
func DefaultMeshConfig() *MeshConfig {
	return &MeshConfig{
		PeerRefreshDelay: 30 * time.Second,
		MaxHops:          3,
		MessageCacheSize: 1000,
		MessageCacheTTL:  5 * time.Minute,
	}
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

// MeshPublisher publishes messages in a mesh network
type MeshPublisher interface {
	Publisher
	JoinMesh(ctx context.Context) error
	LeaveMesh() error
	AnnounceTopics(topics []string) error
	GetConnectedPeers() []string
}

// Subscriber receives messages from publishers
type Subscriber interface {
	Subscribe(ctx context.Context, topics []string, handler MessageHandler) error
	Close() error
}

// PeerConnection represents a connection to a mesh peer
type PeerConnection struct {
	NodeID      string
	ServiceName string
	Conn        net.Conn
	Topics      []string
	LastSeen    time.Time
}

// TopicAnnouncement announces available topics to peers
type TopicAnnouncement struct {
	NodeID    string    `json:"node_id"`
	Topics    []string  `json:"topics"`
	Timestamp time.Time `json:"timestamp"`
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
