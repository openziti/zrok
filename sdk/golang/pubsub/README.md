# OpenZiti Pub/Sub Package

A simple, reliable pub/sub implementation built on OpenZiti with mesh networking support.

## Features

- **OpenZiti Native**: Uses OpenZiti services for secure, overlay network communication
- **Mesh Architecture**: Multiple publishers form a self-organizing mesh for fault tolerance
- **Addressable Terminators**: Uses OpenZiti addressable terminators for peer discovery
- **Auto-Reconnect**: Automatic reconnection with configurable delays and retry limits
- **Topic Filtering**: Subscribe to specific topics or all messages
- **Message Deduplication**: Prevents message loops in mesh networks
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

### Mesh Publisher

```go
package main

import (
    "context"
    "time"
    "github.com/openziti/zrok/sdk/golang/pubsub"
    "github.com/sirupsen/logrus"
)

func main() {
    // configure mesh publisher
    cfg := &pubsub.MeshConfig{
        ServiceName:       "zrok-pubsub-mesh",
        IdentityPath:      "/path/to/identity.json",
        NodeID:           "controller-1",
        PeerRefreshDelay: 30 * time.Second,
        MaxHops:          3,
        MessageCacheSize: 1000,
        MessageCacheTTL:  5 * time.Minute,
    }

    // create mesh publisher
    pub, err := pubsub.NewMeshPublisher(cfg)
    if err != nil {
        logrus.Fatalf("failed to create mesh publisher: %v", err)
    }
    defer pub.Close()

    // join mesh network
    ctx := context.Background()
    if err := pub.JoinMesh(ctx); err != nil {
        logrus.Fatalf("failed to join mesh: %v", err)
    }

    // announce available topics
    pub.AnnounceTopics([]string{"frontend", "metrics", "events"})

    // publish messages (will reach both local subscribers and mesh peers)
    msg := pubsub.NewMessage("hostname_update", "frontend", map[string]any{
        "operation": "ADD",
        "hostname":  "api.example.com",
        "service":   "zrok-service-abc123",
    })

    if err := pub.Publish(ctx, msg); err != nil {
        logrus.Errorf("failed to publish: %v", err)
    }

    logrus.Infof("connected to %d mesh peers", len(pub.GetConnectedPeers()))
}
```

### Mesh Subscriber

```go
package main

import (
    "context"
    "time"
    "github.com/openziti/zrok/sdk/golang/pubsub"
    "github.com/sirupsen/logrus"
)

func main() {
    // configure subscriber (automatically connects to any mesh publisher)
    cfg := &pubsub.SubscriberConfig{
        ServiceName:    "zrok-pubsub-mesh",
        IdentityPath:   "/path/to/subscriber.json",
        ReconnectDelay: 5 * time.Second,
        MaxReconnects:  -1,
    }

    // create mesh-aware subscriber
    sub, err := pubsub.NewMeshSubscriber(cfg, "frontend-1")
    if err != nil {
        logrus.Fatalf("failed to create subscriber: %v", err)
    }
    defer sub.Close()

    // handler automatically deduplicates mesh messages
    handler := func(msg *pubsub.Message) error {
        logrus.Infof("received %s from %s (hops: %d)", 
            msg.Type, msg.OriginNode, msg.HopCount)
        return nil
    }

    ctx := context.Background()
    if err := sub.Subscribe(ctx, []string{"frontend"}, handler); err != nil {
        logrus.Errorf("subscription failed: %v", err)
    }
}
```

## Discovery Mechanism

Peers discover each other through:

1. **Ziti Terminator Query**: Direct query to ziti controller for service terminators
2. **Connection-based Discovery**: Fallback discovery protocol through service connections
3. **Periodic Refresh**: Automatic discovery of new peers every 30 seconds (configurable)

When a new publisher joins:
- It binds the service with its unique node ID as addressable terminator
- Existing peers discover it during their next periodic refresh (≤30s)
- New connections are established automatically
- Mesh forms organically without configuration

## Use Cases

Perfect for distributed systems where:
- Multiple controllers need to coordinate without single points of failure
- Frontend instances need real-time updates from any available controller
- System should gracefully handle controller failures
- Messages need to reach all nodes regardless of which controller publishes