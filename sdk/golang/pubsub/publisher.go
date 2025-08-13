package pubsub

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type publisher struct {
	cfg      *PublisherConfig
	zCtx     ziti.Context
	listener net.Listener
	clients  map[string]net.Conn
	mutex    sync.RWMutex
	done     chan struct{}
}

// NewPublisher creates a new publisher that listens on OpenZiti
func NewPublisher(cfg *PublisherConfig) (Publisher, error) {
	zCfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load ziti config")
	}

	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ziti context")
	}

	listener, err := zCtx.Listen(cfg.ServiceName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to listen on service '%s'", cfg.ServiceName)
	}

	p := &publisher{
		cfg:      cfg,
		zCtx:     zCtx,
		listener: listener,
		clients:  make(map[string]net.Conn),
		done:     make(chan struct{}),
	}

	go p.acceptConnections()
	logrus.Infof("publisher listening on service '%s'", cfg.ServiceName)
	return p, nil
}

func (p *publisher) acceptConnections() {
	for {
		select {
		case <-p.done:
			return
		default:
			conn, err := p.listener.Accept()
			if err != nil {
				select {
				case <-p.done:
					return
				default:
					logrus.Errorf("failed to accept connection: %v", err)
					continue
				}
			}

			clientID := uuid.New().String()
			p.mutex.Lock()
			p.clients[clientID] = conn
			p.mutex.Unlock()

			logrus.Debugf("client connected: %s", clientID)
			go p.handleClient(clientID, conn)
		}
	}
}

func (p *publisher) handleClient(clientID string, conn net.Conn) {
	defer func() {
		p.mutex.Lock()
		delete(p.clients, clientID)
		p.mutex.Unlock()
		conn.Close()
		logrus.Debugf("client disconnected: %s", clientID)
	}()

	// keep connection alive, handle client disconnection
	buffer := make([]byte, 1)
	for {
		select {
		case <-p.done:
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

func (p *publisher) Publish(ctx context.Context, msg *Message) error {
	data, err := msg.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}

	p.mutex.RLock()
	clients := make([]net.Conn, 0, len(p.clients))
	for _, conn := range p.clients {
		clients = append(clients, conn)
	}
	p.mutex.RUnlock()

	var publishErr error
	for _, conn := range clients {
		if _, err := conn.Write(append(data, '\n')); err != nil {
			logrus.Errorf("failed to write to client: %v", err)
			publishErr = err
		}
	}

	logrus.Debugf("published message to %d clients: %s", len(clients), msg.Type)
	return publishErr
}

func (p *publisher) Close() error {
	close(p.done)

	p.mutex.Lock()
	for _, conn := range p.clients {
		conn.Close()
	}
	p.mutex.Unlock()

	if p.listener != nil {
		p.listener.Close()
	}

	return nil
}
