package subordinate

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	MessageKey   = "msg"
	RawMessage   = "raw"
	BootMessage  = "boot"
	ErrorMessage = "error"
)

type Message map[string]interface{}

type MessageHandler struct {
	BootHandler      func(string, Message)
	MessageHandler   func(Message)
	MalformedHandler func(Message)
	BootComplete     chan struct{}
	readBuffer       bytes.Buffer
	booted           bool
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		BootComplete: make(chan struct{}),
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
		logrus.Debugf("line: '%v'", line)
		msg := make(map[string]interface{})
		if !h.booted {
			if line[0] == '{' {
				if err := json.Unmarshal([]byte(line), &msg); err == nil {
					if v, found := msg[MessageKey]; found {
						if vStr, ok := v.(string); ok {
							if vStr == BootMessage || vStr == ErrorMessage {
								h.BootHandler(vStr, msg)
								h.booted = true
								close(h.BootComplete)
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
