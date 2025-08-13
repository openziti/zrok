# OpenZiti Pub/Sub Package

A simple, reliable pub/sub implementation built on OpenZiti.

## Features

- **OpenZiti Native**: Uses OpenZiti services for secure, overlay network communication
- **Auto-Reconnect**: Automatic reconnection with configurable delays and retry limits
- **Topic Filtering**: Subscribe to specific topics or all messages
- **Simple API**: Clean interfaces for both publishers and subscribers
- **Reliability**: Built-in error handling, connection monitoring, and graceful degradation

## Quick Start

### Publisher

```go
package main

import (
    "context"
    "fmt"
    "github.com/openziti/zrok/sdk/golang/pubsub"
    "github.com/sirupsen/logrus"
)

func main() {
    // configure the publisher
    cfg := &pubsub.PublisherConfig{
        ServiceName:  "zrok-pubsub-service",
        IdentityPath: "/path/to/publisher.json",
    }

    // create publisher
    pub, err := pubsub.NewPublisher(cfg)
    if err != nil {
        logrus.Fatalf("failed to create publisher: %v", err)
    }
    defer pub.Close()

    // publish dynamic hostname mapping
    msg := pubsub.NewMessage("hostname_update", "frontend", map[string]any{
        "operation": "ADD",
        "hostname":  "api.example.com",
        "service":   "zrok-service-abc123",
        "ttl":       3600,
    })

    ctx := context.Background()
    if err := pub.Publish(ctx, msg); err != nil {
        logrus.Errorf("failed to publish: %v", err)
    }

    fmt.Println("message published successfully")
}
```

### Subscriber (Frontend Side)

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/openziti/zrok/sdk/golang/pubsub"
    "github.com/sirupsen/logrus"
)

func main() {
    // configure the subscriber with auto-reconnect
    cfg := &pubsub.SubscriberConfig{
        ServiceName:    "zrok-pubsub-service",
        IdentityPath:   "/path/to/subscriber.json",
        ReconnectDelay: 5 * time.Second,
        MaxReconnects:  -1, // infinite reconnects
        MessageTimeout: 30 * time.Second,
    }

    // create subscriber
    sub, err := pubsub.NewSubscriber(cfg)
    if err != nil {
        logrus.Fatalf("failed to create subscriber: %v", err)
    }
    defer sub.Close()

    // define message handler
    handler := func(msg *pubsub.Message) error {
        fmt.Printf("received %s message on topic %s\n", msg.Type, msg.Topic)

        // handle hostname updates for frontend
        if msg.Type == "hostname_update" && msg.Topic == "frontend" {
            operation := msg.Data["operation"].(string)
            hostname := msg.Data["hostname"].(string)
            service := msg.Data["service"].(string)

            switch operation {
            case "ADD":
                fmt.Printf("adding mapping: %s -> %s\n", hostname, service)
                // add to cache here
            case "DELETE":
                fmt.Printf("removing mapping: %s\n", hostname)
                // remove from cache here
            case "UPDATE":
                fmt.Printf("updating mapping: %s -> %s\n", hostname, service)
                // update cache here
            }
        }

        return nil
    }

    // subscribe to frontend topic
    ctx := context.Background()
    topics := []string{"frontend"}

    if err := sub.Subscribe(ctx, topics, handler); err != nil {
        logrus.Errorf("subscription failed: %v", err)
    }

    fmt.Println("subscriber started with auto-reconnect")
}
```

## Message Format

```json
{
    "id": "abc123def456",
    "type": "hostname_update", 
    "topic": "frontend",
    "data": {
        "operation": "ADD",
        "hostname": "api.example.com",
        "service": "zrok-service-123",
        "ttl": 3600
    },
    "timestamp": "2025-08-13T10:30:00Z"
}
```

## Configuration

### SubscriberConfig Options

- `ReconnectDelay`: Time to wait between reconnection attempts
- `MaxReconnects`: Maximum reconnection attempts (-1 = infinite)
- `MessageTimeout`: Timeout for message processing

## Use Cases

Perfect for the dynamic frontend system where:
- Controller publishes hostname → service mappings
- Multiple frontend instances subscribe to updates
- Automatic failover when connections drop
- Real-time configuration without restarts