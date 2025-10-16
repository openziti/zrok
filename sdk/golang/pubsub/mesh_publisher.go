package pubsub

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/pkg/errors"
)

type meshPublisher struct {
	cfg          *MeshConfig
	node         *meshNode
	listener     net.Listener
	localClients map[string]net.Conn
	clientMutex  sync.RWMutex
	topics       []string
	topicMutex   sync.RWMutex
	done         chan struct{}
}

// NewMeshPublisher creates a new mesh-enabled publisher
func NewMeshPublisher(cfg *MeshConfig) (MeshPublisher, error) {
	if cfg == nil {
		return nil, errors.New("mesh config is required")
	}

	if cfg.NodeID == "" {
		return nil, errors.New("node ID is required")
	}

	// apply defaults
	if cfg.PeerRefreshDelay == 0 {
		defaults := DefaultMeshConfig()
		cfg.PeerRefreshDelay = defaults.PeerRefreshDelay
		cfg.MaxHops = defaults.MaxHops
		cfg.MessageCacheSize = defaults.MessageCacheSize
		cfg.MessageCacheTTL = defaults.MessageCacheTTL
	}

	// create mesh node for peer management
	node, err := newMeshNode(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mesh node")
	}

	// create listener for local subscribers using addressable terminator
	listener, err := node.zCtx.ListenWithOptions(cfg.ServiceName, &ziti.ListenOptions{
		// note: actual addressable terminator configuration may vary
		// this is conceptual and would need ziti SDK specifics
		Identity: cfg.NodeID,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to listen on service '%s' with identity '%s'", cfg.ServiceName, cfg.NodeID)
	}

	mp := &meshPublisher{
		cfg:          cfg,
		node:         node,
		listener:     listener,
		localClients: make(map[string]net.Conn),
		done:         make(chan struct{}),
	}

	// start accepting local client connections
	go mp.acceptLocalConnections()

	dl.Infof("mesh publisher listening on service '%s' with node ID '%s'", cfg.ServiceName, cfg.NodeID)
	return mp, nil
}

func (mp *meshPublisher) acceptLocalConnections() {
	for {
		select {
		case <-mp.done:
			return
		default:
			conn, err := mp.listener.Accept()
			if err != nil {
				select {
				case <-mp.done:
					return
				default:
					dl.Errorf("failed to accept local connection: %v", err)
					continue
				}
			}

			clientID := fmt.Sprintf("local-%d", time.Now().UnixNano())
			mp.clientMutex.Lock()
			mp.localClients[clientID] = conn
			mp.clientMutex.Unlock()

			dl.Debugf("local client connected: %s", clientID)
			go mp.handleLocalClient(clientID, conn)
		}
	}
}

func (mp *meshPublisher) handleLocalClient(clientID string, conn net.Conn) {
	defer func() {
		mp.clientMutex.Lock()
		delete(mp.localClients, clientID)
		mp.clientMutex.Unlock()
		conn.Close()
		dl.Debugf("local client disconnected: %s", clientID)
	}()

	// handle both subscriber connections and discovery requests
	scanner := bufio.NewScanner(conn)
	scanner.Scan() // read first message to determine connection type

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "" && strings.Contains(firstLine, "peer_discovery") {
		// handle discovery request
		mp.handleDiscoveryRequest(conn, firstLine)
		return
	}

	// regular subscriber connection - keep alive
	buffer := make([]byte, 1)
	for {
		select {
		case <-mp.done:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			_, err := conn.Read(buffer)
			if err != nil {
				return // client disconnected
			}
		}
	}
}

func (mp *meshPublisher) handleDiscoveryRequest(conn net.Conn, request string) {
	var req map[string]interface{}
	if err := json.Unmarshal([]byte(request), &req); err != nil {
		dl.Debugf("invalid discovery request: %v", err)
		return
	}

	if action, ok := req["action"].(string); ok && action == "request_peers" {
		// respond with list of known peers
		peers := mp.node.getConnectedPeers()
		peers = append(peers, mp.cfg.NodeID) // include self

		response := map[string]interface{}{
			"type":  "peer_discovery_response",
			"peers": peers,
		}

		data, err := json.Marshal(response)
		if err == nil {
			conn.Write(append(data, '\n'))
		}

		dl.Debugf("responded to peer discovery request with %d peers", len(peers))
	}
}

// JoinMesh connects to the mesh network
func (mp *meshPublisher) JoinMesh(ctx context.Context) error {
	return mp.node.joinMesh(ctx)
}

// LeaveMesh disconnects from the mesh network
func (mp *meshPublisher) LeaveMesh() error {
	return mp.node.leaveMesh()
}

// AnnounceTopics announces available topics to mesh peers
func (mp *meshPublisher) AnnounceTopics(topics []string) error {
	mp.topicMutex.Lock()
	mp.topics = make([]string, len(topics))
	copy(mp.topics, topics)
	mp.topicMutex.Unlock()

	return mp.node.announceTopics(topics)
}

// GetConnectedPeers returns list of connected peer node IDs
func (mp *meshPublisher) GetConnectedPeers() []string {
	return mp.node.getConnectedPeers()
}

// Publish publishes a message to both local subscribers and mesh peers
func (mp *meshPublisher) Publish(ctx context.Context, msg *Message) error {
	// set origin node if not already set
	if msg.OriginNode == "" {
		msg.OriginNode = mp.cfg.NodeID
	}

	// prevent duplicate processing
	if mp.node.messageCache.isDuplicate(msg.Id) {
		return nil
	}

	// marshal message once
	data, err := msg.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}

	var publishErr error

	// publish to local subscribers
	if err := mp.publishToLocalClients(data); err != nil {
		publishErr = err
	}

	// relay to mesh peers (if not a relayed message)
	if msg.HopCount == 0 && msg.OriginNode == mp.cfg.NodeID {
		if err := mp.node.relayMessage(msg); err != nil {
			dl.Errorf("failed to relay message to mesh: %v", err)
			if publishErr == nil {
				publishErr = err
			}
		}
	}

	dl.Debugf("published message: %s (topic: %s, origin: %s, hops: %d)",
		msg.Id, msg.Topic, msg.OriginNode, msg.HopCount)

	return publishErr
}

func (mp *meshPublisher) publishToLocalClients(data []byte) error {
	mp.clientMutex.RLock()
	clients := make([]net.Conn, 0, len(mp.localClients))
	for _, conn := range mp.localClients {
		clients = append(clients, conn)
	}
	mp.clientMutex.RUnlock()

	var publishErr error
	for _, conn := range clients {
		if _, err := conn.Write(append(data, '\n')); err != nil {
			dl.Errorf("failed to write to local client: %v", err)
			publishErr = err
		}
	}

	dl.Debugf("published to %d local clients", len(clients))
	return publishErr
}

// Close shuts down the mesh publisher
func (mp *meshPublisher) Close() error {
	close(mp.done)

	// leave mesh
	if err := mp.node.leaveMesh(); err != nil {
		dl.Errorf("error leaving mesh: %v", err)
	}

	// close local client connections
	mp.clientMutex.Lock()
	for _, conn := range mp.localClients {
		conn.Close()
	}
	mp.clientMutex.Unlock()

	// close listener
	if mp.listener != nil {
		mp.listener.Close()
	}

	return nil
}
