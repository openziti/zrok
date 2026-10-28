package subordinate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/michaelquigley/df/dd"
)

const (
	MessageKey   = "msg"
	RawMessage   = "raw"
	BootMessage  = "boot"
	ErrorMessage = "error"
	MaxLineSize  = 1024 * 1024
)

type Message map[string]interface{}

type MessageHandler struct {
	BootHandler      func(string, Message)
	MessageHandler   func(Message)
	MalformedHandler func(Message)
	BootComplete     chan struct{}
	mu               sync.Mutex
	pending          []byte
	booted           bool
	closed           bool
	err              error
}

type messageHeader struct {
	Msg string `dd:"msg,+required"`
}

func NewMessageHandler() *MessageHandler { return &MessageHandler{BootComplete: make(chan struct{})} }

// Tail drains complete lines and retains only the trailing fragment, with a fixed memory bound.
func (h *MessageHandler) Tail(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || h.err != nil {
		return
	}
	for len(data) > 0 {
		end := bytes.IndexByte(data, '\n')
		count := len(data)
		if end >= 0 {
			count = end
		}
		if len(h.pending)+count > MaxLineSize {
			h.fail(fmt.Errorf("subordinate protocol line exceeds '%v' bytes", MaxLineSize))
			h.pending = nil
			return
		}
		h.pending = append(h.pending, data[:count]...)
		if end < 0 {
			return
		}
		h.line(bytes.TrimSpace(h.pending))
		h.pending = h.pending[:0]
		if h.err != nil {
			return
		}
		data = data[end+1:]
	}
}

// Close reports EOF before boot as startup failure; an unterminated boot line is not readiness.
func (h *MessageHandler) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	h.pending = nil
	if !h.booted && h.err == nil {
		h.fail(fmt.Errorf("subordinate output closed before boot: %w", io.ErrUnexpectedEOF))
	}
}

func (h *MessageHandler) Err() error { h.mu.Lock(); defer h.mu.Unlock(); return h.err }

func (h *MessageHandler) fail(err error) {
	h.err = err
	msg := Message{MessageKey: ErrorMessage, ErrorMessage: err.Error()}
	if !h.booted {
		h.boot(ErrorMessage, msg)
	} else {
		h.message(msg)
	}
}

func (h *MessageHandler) line(line []byte) {
	if len(line) == 0 {
		return
	}
	msg := Message{}
	// this is an opaque JSON map bridge; project field binding goes through dd below and in boot handlers.
	if line[0] != '{' || json.Unmarshal(line, &msg) != nil {
		h.message(Message{MessageKey: RawMessage, RawMessage: string(line)})
		return
	}
	header, err := dd.New[messageHeader](msg)
	if err != nil || header.Msg == "" {
		if h.MalformedHandler != nil {
			h.MalformedHandler(msg)
		}
		return
	}
	if !h.booted && (header.Msg == BootMessage || header.Msg == ErrorMessage) {
		if header.Msg == ErrorMessage {
			if text, ok := msg[ErrorMessage].(string); !ok || text == "" {
				h.fail(errors.New("subordinate reported an error without a message"))
				return
			}
		}
		h.boot(header.Msg, msg)
	} else {
		h.message(msg)
	}
}

func (h *MessageHandler) boot(kind string, msg Message) {
	if h.BootHandler != nil {
		h.BootHandler(kind, msg)
	}
	h.booted = true
	close(h.BootComplete)
}

func (h *MessageHandler) message(msg Message) {
	if h.MessageHandler != nil {
		h.MessageHandler(msg)
	}
}
