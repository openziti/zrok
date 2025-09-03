package pubsub

import (
	"bufio"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// meshNode handles peer discovery and mesh coordination
type meshNode struct {
	cfg          *MeshConfig
	zCtx         ziti.Context
	peers        map[string]*PeerConnection
	messageCache *messageDeduplicator
	mutex        sync.RWMutex
	done         chan struct{}
	peerRefresh  *time.Ticker
}

// messageDeduplicator prevents message loops in the mesh
type messageDeduplicator struct {
	seen   map[string]time.Time
	maxAge time.Duration
	mutex  sync.RWMutex
}

func newMessageDeduplicator(maxAge time.Duration, cacheSize int) *messageDeduplicator {
	md := &messageDeduplicator{
		seen:   make(map[string]time.Time, cacheSize),
		maxAge: maxAge,
	}

	// periodic cleanup of old entries
	go func() {
		ticker := time.NewTicker(maxAge / 2)
		defer ticker.Stop()

		for range ticker.C {
			md.cleanup()
		}
	}()

	return md
}

func (md *messageDeduplicator) isDuplicate(messageID string) bool {
	md.mutex.RLock()
	_, seen := md.seen[messageID]
	md.mutex.RUnlock()

	if seen {
		return true
	}

	md.mutex.Lock()
	md.seen[messageID] = time.Now()
	md.mutex.Unlock()

	return false
}

func (md *messageDeduplicator) cleanup() {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	cutoff := time.Now().Add(-md.maxAge)
	for id, timestamp := range md.seen {
		if timestamp.Before(cutoff) {
			delete(md.seen, id)
		}
	}
}

func newMeshNode(cfg *MeshConfig) (*meshNode, error) {
	zCfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load ziti config")
	}

	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ziti context")
	}

	return &meshNode{
		cfg:          cfg,
		zCtx:         zCtx,
		peers:        make(map[string]*PeerConnection),
		messageCache: newMessageDeduplicator(cfg.MessageCacheTTL, cfg.MessageCacheSize),
		done:         make(chan struct{}),
	}, nil
}

// joinMesh discovers peers and establishes connections
func (mn *meshNode) joinMesh(ctx context.Context) error {
	// start peer discovery and refresh
	mn.peerRefresh = time.NewTicker(mn.cfg.PeerRefreshDelay)
	go mn.managePeers(ctx)

	// initial peer discovery
	if err := mn.discoverAndConnectPeers(ctx); err != nil {
		logrus.Warnf("initial peer discovery failed: %v", err)
	}

	return nil
}

func (mn *meshNode) leaveMesh() error {
	close(mn.done)

	if mn.peerRefresh != nil {
		mn.peerRefresh.Stop()
	}

	// close all peer connections
	mn.mutex.Lock()
	for _, peer := range mn.peers {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
	}
	mn.peers = make(map[string]*PeerConnection)
	mn.mutex.Unlock()

	return nil
}

// discoverPeers finds available peers using ziti service introspection
func (mn *meshNode) discoverPeers() ([]string, error) {
	logrus.Debugf("discovering peers for service '%s'", mn.cfg.ServiceName)

	// approach 1: query ziti controller for service terminators
	// note: this requires appropriate ziti API access
	peers, err := mn.queryZitiTerminators()
	if err == nil && len(peers) > 0 {
		return peers, nil
	}

	// approach 2: attempt connections to discover active peers
	// use a discovery service pattern if direct terminator query isn't available
	return mn.discoverByConnection()
}

// queryZitiTerminators queries the ziti controller for available terminators
func (mn *meshNode) queryZitiTerminators() ([]string, error) {
	// note: actual implementation depends on ziti SDK capabilities
	// this would use ziti controller API to list service terminators

	// conceptual implementation:
	// terminators, err := mn.zCtx.GetServiceTerminators(mn.cfg.ServiceName)
	// if err != nil {
	//     return nil, err
	// }
	//
	// var peers []string
	// for _, terminator := range terminators {
	//     if terminator.Identity != mn.cfg.NodeID {
	//         peers = append(peers, terminator.Identity)
	//     }
	// }
	// return peers, nil

	return nil, errors.New("ziti terminator query not implemented")
}

// discoverByConnection attempts to discover peers through connection attempts
func (mn *meshNode) discoverByConnection() ([]string, error) {
	// fallback: attempt to connect to the service without specifying identity
	// and use a discovery handshake protocol

	conn, err := mn.zCtx.DialWithOptions(mn.cfg.ServiceName, &ziti.DialOptions{
		ConnectTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect for peer discovery")
	}
	defer conn.Close()

	// send discovery request
	discoveryReq := map[string]interface{}{
		"type":    "peer_discovery",
		"node_id": mn.cfg.NodeID,
		"action":  "request_peers",
	}

	data, err := json.Marshal(discoveryReq)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return nil, err
	}

	// read response with timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var response map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			return nil, err
		}

		if peers, ok := response["peers"].([]interface{}); ok {
			var peerList []string
			for _, peer := range peers {
				if peerStr, ok := peer.(string); ok && peerStr != mn.cfg.NodeID {
					peerList = append(peerList, peerStr)
				}
			}
			return peerList, nil
		}
	}

	return nil, errors.New("no valid peer discovery response")
}

func (mn *meshNode) discoverAndConnectPeers(ctx context.Context) error {
	peerIDs, err := mn.discoverPeers()
	if err != nil {
		return errors.Wrap(err, "failed to discover peers")
	}

	mn.mutex.Lock()
	existingPeers := make(map[string]bool)
	for peerID := range mn.peers {
		existingPeers[peerID] = true
	}
	mn.mutex.Unlock()

	// connect to new peers
	for _, peerID := range peerIDs {
		if peerID == mn.cfg.NodeID {
			continue // skip self
		}

		if existingPeers[peerID] {
			continue // already connected
		}

		go mn.connectToPeer(ctx, peerID)
	}

	return nil
}

func (mn *meshNode) connectToPeer(ctx context.Context, peerNodeID string) {
	logrus.Debugf("connecting to peer: %s", peerNodeID)

	// use addressable terminator to connect to specific peer
	conn, err := mn.zCtx.DialWithOptions(mn.cfg.ServiceName, &ziti.DialOptions{
		ConnectTimeout: 30 * time.Second,
		// note: actual addressable terminator syntax may vary
		// this is conceptual - would need ziti SDK specifics
	})
	if err != nil {
		logrus.Warnf("failed to connect to peer '%s': %v", peerNodeID, err)
		return
	}

	peer := &PeerConnection{
		NodeID:      peerNodeID,
		ServiceName: mn.cfg.ServiceName,
		Conn:        conn,
		LastSeen:    time.Now(),
	}

	mn.mutex.Lock()
	mn.peers[peerNodeID] = peer
	mn.mutex.Unlock()

	logrus.Infof("connected to peer: %s", peerNodeID)

	// handle peer communication
	go mn.handlePeerConnection(peer)
}

func (mn *meshNode) handlePeerConnection(peer *PeerConnection) {
	defer func() {
		mn.mutex.Lock()
		delete(mn.peers, peer.NodeID)
		mn.mutex.Unlock()

		if peer.Conn != nil {
			peer.Conn.Close()
		}

		logrus.Debugf("disconnected from peer: %s", peer.NodeID)
	}()

	scanner := bufio.NewScanner(peer.Conn)
	for scanner.Scan() {
		select {
		case <-mn.done:
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// handle different message types from peer
		if strings.HasPrefix(line, "{") {
			// json message - could be pub/sub message or control message
			mn.handlePeerMessage(peer, line)
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Debugf("peer connection error for %s: %v", peer.NodeID, err)
	}
}

func (mn *meshNode) handlePeerMessage(peer *PeerConnection, line string) {
	// try to parse as topic announcement first
	var announcement TopicAnnouncement
	if err := json.Unmarshal([]byte(line), &announcement); err == nil && announcement.NodeID != "" {
		mn.handleTopicAnnouncement(peer, &announcement)
		return
	}

	// try to parse as pub/sub message
	var msg Message
	if err := json.Unmarshal([]byte(line), &msg); err == nil && msg.Id != "" {
		mn.handleRelayedMessage(&msg)
		return
	}

	logrus.Debugf("unknown message from peer %s: %s", peer.NodeID, line)
}

func (mn *meshNode) handleTopicAnnouncement(peer *PeerConnection, announcement *TopicAnnouncement) {
	mn.mutex.Lock()
	if existingPeer, exists := mn.peers[announcement.NodeID]; exists {
		existingPeer.Topics = announcement.Topics
		existingPeer.LastSeen = time.Now()
	}
	mn.mutex.Unlock()

	logrus.Debugf("received topic announcement from %s: %v", announcement.NodeID, announcement.Topics)
}

func (mn *meshNode) handleRelayedMessage(msg *Message) {
	// check for duplicates and hop limits
	if mn.messageCache.isDuplicate(msg.Id) {
		return
	}

	if msg.HopCount >= mn.cfg.MaxHops {
		logrus.Debugf("dropping message due to hop limit: %s", msg.Id)
		return
	}

	// this would be handled by the mesh publisher's relay logic
	logrus.Debugf("received relayed message: %s (topic: %s, hops: %d)", msg.Id, msg.Topic, msg.HopCount)
}

func (mn *meshNode) managePeers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-mn.done:
			return
		case <-mn.peerRefresh.C:
			if err := mn.discoverAndConnectPeers(ctx); err != nil {
				logrus.Debugf("peer discovery error: %v", err)
			}
			mn.cleanupStaleConnections()
		}
	}
}

func (mn *meshNode) cleanupStaleConnections() {
	mn.mutex.Lock()
	defer mn.mutex.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for peerID, peer := range mn.peers {
		if peer.LastSeen.Before(cutoff) {
			logrus.Debugf("removing stale peer connection: %s", peerID)
			if peer.Conn != nil {
				peer.Conn.Close()
			}
			delete(mn.peers, peerID)
		}
	}
}

// relayMessage sends a message to all connected peers
func (mn *meshNode) relayMessage(msg *Message) error {
	// increment hop count
	msg.HopCount++

	if msg.HopCount >= mn.cfg.MaxHops {
		return nil // don't relay beyond hop limit
	}

	data, err := msg.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal message for relay")
	}

	mn.mutex.RLock()
	peers := make([]*PeerConnection, 0, len(mn.peers))
	for _, peer := range mn.peers {
		peers = append(peers, peer)
	}
	mn.mutex.RUnlock()

	// send to all peers
	for _, peer := range peers {
		if peer.Conn != nil {
			go func(p *PeerConnection) {
				if _, err := p.Conn.Write(append(data, '\n')); err != nil {
					logrus.Debugf("failed to relay message to peer %s: %v", p.NodeID, err)
				}
			}(peer)
		}
	}

	logrus.Debugf("relayed message to %d peers: %s", len(peers), msg.Id)
	return nil
}

// announceTopics broadcasts available topics to peers
func (mn *meshNode) announceTopics(topics []string) error {
	announcement := &TopicAnnouncement{
		NodeID:    mn.cfg.NodeID,
		Topics:    topics,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(announcement)
	if err != nil {
		return errors.Wrap(err, "failed to marshal topic announcement")
	}

	mn.mutex.RLock()
	peers := make([]*PeerConnection, 0, len(mn.peers))
	for _, peer := range mn.peers {
		peers = append(peers, peer)
	}
	mn.mutex.RUnlock()

	for _, peer := range peers {
		if peer.Conn != nil {
			go func(p *PeerConnection) {
				if _, err := p.Conn.Write(append(data, '\n')); err != nil {
					logrus.Debugf("failed to announce topics to peer %s: %v", p.NodeID, err)
				}
			}(peer)
		}
	}

	return nil
}

// getConnectedPeers returns list of connected peer node IDs
func (mn *meshNode) getConnectedPeers() []string {
	mn.mutex.RLock()
	defer mn.mutex.RUnlock()

	peers := make([]string, 0, len(mn.peers))
	for peerID := range mn.peers {
		peers = append(peers, peerID)
	}

	return peers
}
