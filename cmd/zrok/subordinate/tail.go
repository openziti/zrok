package subordinate

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	MessageKey  = "msg"
	RawMessage  = "raw"
	BootMessage = "boot"
)

type Message map[string]interface{}

type MessageHandler struct {
	BootHandler      func(Message)
	MessageHandler   func(Message)
	MalformedHandler func(Message)
	readBuffer       bytes.Buffer
	booted           bool
	bootComplete     chan struct{}
	bootErr          error
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		bootComplete: make(chan struct{}),
	}
}

func (h *MessageHandler) Tail(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("recovered: %v", r)
		}
	}()

	h.readBuffer.Write(data)
	if line, err := h.readBuffer.ReadString('\n'); err == nil {
		line = strings.Trim(line, "\n \t")
		msg := make(map[string]interface{})
		if !h.booted {
			if line[0] == '{' {
				if err := json.Unmarshal([]byte(line), &msg); err == nil {
					if v, found := msg[MessageKey]; found {
						if vStr, ok := v.(string); ok {
							if vStr == BootMessage {
								h.BootHandler(msg)
								h.booted = true
								close(h.bootComplete)
							} else {
								h.MessageHandler(msg)
							}
						} else {
							h.MalformedHandler(msg)
						}
					} else {
						h.MalformedHandler(msg)
					}
				} else {
					msg[MessageKey] = RawMessage
					msg[RawMessage] = line
					h.MessageHandler(msg)
				}
			} else {
				msg[MessageKey] = RawMessage
				msg[RawMessage] = line
				h.MessageHandler(msg)
			}
		} else {
			if line[0] == '{' {
				if err := json.Unmarshal([]byte(line), &msg); err != nil {
					logrus.Error(line)
				}
			} else {
				msg[MessageKey] = RawMessage
				msg[RawMessage] = line
			}
			h.MessageHandler(msg)
		}
	}
}
