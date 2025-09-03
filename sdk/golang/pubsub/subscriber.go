package pubsub

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type subscriber struct {
	cfg            *SubscriberConfig
	zCtx           ziti.Context
	conn           net.Conn
	topics         []string
	handler        MessageHandler
	mutex          sync.RWMutex
	done           chan struct{}
	reconnectCount int
	connected      bool
}

// NewSubscriber creates a new subscriber that connects to OpenZiti service
func NewSubscriber(cfg *SubscriberConfig) (Subscriber, error) {
	if cfg == nil {
		cfg = DefaultSubscriberConfig()
	}

	zCfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load ziti config")
	}

	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ziti context")
	}

	s := &subscriber{
		cfg:  cfg,
		zCtx: zCtx,
		done: make(chan struct{}),
	}

	return s, nil
}

func (s *subscriber) Subscribe(ctx context.Context, topics []string, handler MessageHandler) error {
	s.mutex.Lock()
	s.topics = topics
	s.handler = handler
	s.mutex.Unlock()

	return s.connect(ctx)
}

func (s *subscriber) connect(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		default:
		}

		// try to connect to any available mesh publisher
		// note: in a mesh setup, we connect to any available terminator
		conn, err := s.zCtx.DialWithOptions(s.cfg.ServiceName, &ziti.DialOptions{
			ConnectTimeout: 30 * time.Second,
			// in mesh mode, we don't specify a particular identity
			// and let ziti route us to any available terminator
		})
		if err != nil {
			if s.shouldReconnect() {
				logrus.Warnf("failed to connect to pubsub service '%s', retrying in %v: %v",
					s.cfg.ServiceName, s.cfg.ReconnectDelay, err)
				s.reconnectCount++

				select {
				case <-time.After(s.cfg.ReconnectDelay):
					continue
				case <-ctx.Done():
					return ctx.Err()
				case <-s.done:
					return nil
				}
			}
			return errors.Wrapf(err, "failed to connect to pubsub service '%s'", s.cfg.ServiceName)
		}

		s.mutex.Lock()
		s.conn = conn
		s.connected = true
		s.reconnectCount = 0
		s.mutex.Unlock()

		logrus.Infof("connected to pubsub service '%s'", s.cfg.ServiceName)

		// start message reading loop
		go s.readMessages(ctx)

		// wait for disconnection
		s.waitForDisconnection()

		// if we reach here, connection was lost
		s.mutex.Lock()
		s.connected = false
		if s.conn != nil {
			s.conn.Close()
			s.conn = nil
		}
		s.mutex.Unlock()

		// check if we should reconnect
		if !s.shouldReconnect() {
			return errors.New("max reconnection attempts reached")
		}

		logrus.Warnf("connection lost, attempting to reconnect in %v", s.cfg.ReconnectDelay)
		select {
		case <-time.After(s.cfg.ReconnectDelay):
			continue
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		}
	}
}

func (s *subscriber) shouldReconnect() bool {
	return s.cfg.MaxReconnects < 0 || s.reconnectCount < s.cfg.MaxReconnects
}

func (s *subscriber) waitForDisconnection() {
	s.mutex.RLock()
	conn := s.conn
	s.mutex.RUnlock()

	if conn == nil {
		return
	}

	// keep connection alive with periodic pings
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			// send keep-alive ping
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if _, err := conn.Write([]byte("\n")); err != nil {
				logrus.Debugf("connection lost during ping: %v", err)
				return
			}
		}
	}
}

func (s *subscriber) readMessages(ctx context.Context) {
	s.mutex.RLock()
	conn := s.conn
	s.mutex.RUnlock()

	if conn == nil {
		return
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // skip empty lines (keep-alive pings)
		}

		msg, err := UnmarshalMessage([]byte(line))
		if err != nil {
			logrus.Errorf("failed to unmarshal message: %v", err)
			continue
		}

		// check if we're interested in this topic
		if s.isInterestedInTopic(msg.Topic) {
			s.mutex.RLock()
			handler := s.handler
			s.mutex.RUnlock()

			if handler != nil {
				go func(m *Message) {
					if err := handler(m); err != nil {
						logrus.Errorf("message handler error: %v", err)
					}
				}(msg)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Debugf("scanner error: %v", err)
	}
}

func (s *subscriber) isInterestedInTopic(topic string) bool {
	s.mutex.RLock()
	topics := s.topics
	s.mutex.RUnlock()

	if len(topics) == 0 {
		return true // subscribe to all topics
	}

	for _, t := range topics {
		if t == topic || t == "*" {
			return true
		}
	}
	return false
}

func (s *subscriber) Close() error {
	close(s.done)

	s.mutex.Lock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.connected = false
	s.mutex.Unlock()

	return nil
}
