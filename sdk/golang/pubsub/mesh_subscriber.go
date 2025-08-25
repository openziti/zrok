package pubsub

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// meshSubscriber extends the basic subscriber for mesh-aware operations
type meshSubscriber struct {
	*subscriber
	nodeID           string
	messageCache     *messageDeduplicator
	processedHandler MessageHandler
}

// NewMeshSubscriber creates a mesh-aware subscriber
func NewMeshSubscriber(cfg *SubscriberConfig, nodeID string) (Subscriber, error) {
	// create base subscriber
	baseSubscriber, err := NewSubscriber(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create base subscriber")
	}

	// create message deduplicator
	messageCache := newMessageDeduplicator(5*time.Minute, 1000)

	ms := &meshSubscriber{
		subscriber:   baseSubscriber.(*subscriber),
		nodeID:       nodeID,
		messageCache: messageCache,
	}

	return ms, nil
}

// Subscribe wraps the base subscribe with mesh-aware message handling
func (ms *meshSubscriber) Subscribe(ctx context.Context, topics []string, handler MessageHandler) error {
	// wrap the handler to add mesh-specific processing
	ms.processedHandler = handler
	wrappedHandler := ms.wrapMessageHandler(handler)

	return ms.subscriber.Subscribe(ctx, topics, wrappedHandler)
}

// wrapMessageHandler adds mesh-specific logic to message processing
func (ms *meshSubscriber) wrapMessageHandler(originalHandler MessageHandler) MessageHandler {
	return func(msg *Message) error {
		// check for duplicate messages in mesh
		if ms.messageCache.isDuplicate(msg.Id) {
			logrus.Debugf("ignoring duplicate message: %s", msg.Id)
			return nil
		}

		// add mesh-specific logging
		if msg.OriginNode != "" && msg.OriginNode != ms.nodeID {
			logrus.Debugf("processing message from mesh peer '%s': %s (topic: %s, hops: %d)", 
				msg.OriginNode, msg.Id, msg.Topic, msg.HopCount)
		} else {
			logrus.Debugf("processing local message: %s (topic: %s)", msg.Id, msg.Topic)
		}

		// call original handler
		return originalHandler(msg)
	}
}

// Close extends base close with mesh-specific cleanup
func (ms *meshSubscriber) Close() error {
	return ms.subscriber.Close()
}