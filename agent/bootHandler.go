package agent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/michaelquigley/df/dd"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

// BootHandlerConfig defines the field mappings for extracting values from boot messages
type BootHandlerConfig struct {
	// validate checks the operation-specific boot contract before applying any field setters.
	validate func(subordinate.Message) error
	// StringFields maps message keys to functions that set string values
	StringFields map[string]func(string)
	// ArrayFields maps message keys to functions that set string array values
	ArrayFields map[string]func([]string)
	// OperationType for logging and debugging purposes
	OperationType string
}

// UnifiedBootHandler provides centralized boot message processing
// eliminating duplication across share and access operations
type UnifiedBootHandler struct {
	config  *BootHandlerConfig
	bootErr *error
}

// NewBootHandler creates a new unified boot handler with the specified configuration
func NewBootHandler(config *BootHandlerConfig, bootErr *error) *UnifiedBootHandler {
	return &UnifiedBootHandler{
		config:  config,
		bootErr: bootErr,
	}
}

// HandleBoot processes boot and error messages from subordinate processes
func (ubh *UnifiedBootHandler) HandleBoot(msgType string, msg subordinate.Message) {
	switch msgType {
	case subordinate.BootMessage:
		ubh.processBootMessage(msg)
	case subordinate.ErrorMessage:
		ubh.processErrorMessage(msg)
	}
}

// HandleMalformed processes malformed messages from subordinate processes
func (ubh *UnifiedBootHandler) HandleMalformed(msg subordinate.Message) {
	dl.Error(msg)
}

// processBootMessage extracts field values from boot messages using configured mappings
func (ubh *UnifiedBootHandler) processBootMessage(msg subordinate.Message) {
	if ubh.config.validate != nil {
		if err := ubh.config.validate(msg); err != nil {
			*ubh.bootErr = err
			return
		}
	}
	// process string fields
	for key, setter := range ubh.config.StringFields {
		if v, found := msg[key]; found {
			if str, ok := v.(string); ok {
				setter(str)
			}
		}
	}

	// process array fields
	for key, setter := range ubh.config.ArrayFields {
		if v, found := msg[key]; found {
			if vArr, ok := v.([]interface{}); ok {
				var stringArr []string
				for _, item := range vArr {
					if str, ok := item.(string); ok {
						stringArr = append(stringArr, str)
					}
				}
				if len(stringArr) > 0 {
					setter(stringArr)
				}
			}
		}
	}
}

// processErrorMessage extracts error information and sets the boot error
func (ubh *UnifiedBootHandler) processErrorMessage(msg subordinate.Message) {
	*ubh.bootErr = errors.New("subordinate reported an error without a message")
	if v, found := msg[subordinate.ErrorMessage]; found {
		if str, ok := v.(string); ok {
			*ubh.bootErr = errors.New(str)
		}
	}
}

// NewShareBootHandler creates a boot handler configured for share operations
func NewShareBootHandler(shr *share, bootErr *error) *UnifiedBootHandler {
	config := &BootHandlerConfig{
		StringFields: map[string]func(string){
			"token": func(v string) { shr.token = v },
			"backend_mode": func(v string) {
				shr.backendMode = sdk.BackendMode(v)
			},
			"share_mode": func(v string) {
				shr.shareMode = sdk.ShareMode(v)
			},
			"target": func(v string) { shr.target = v },
		},
		ArrayFields: map[string]func([]string){
			"frontend_endpoints": func(v []string) {
				shr.frontendEndpoints = v
			},
		},
		OperationType: "share",
	}
	return NewBootHandler(config, bootErr)
}

// NewAccessBootHandler creates a boot handler configured for access operations
func NewAccessBootHandler(acc *access, bootErr *error) *UnifiedBootHandler {
	config := &BootHandlerConfig{
		StringFields: map[string]func(string){
			"frontend_token": func(v string) { acc.frontendToken = v },
			"bind_address":   func(v string) { acc.bindAddress = v },
		},
		OperationType: "access",
	}
	return NewBootHandler(config, bootErr)
}

// publicWorkerBoot binds the public worker's readiness fields; endpoints remain owned by the parent.
type publicWorkerBoot struct {
	Token string `dd:"token,+required"`
}

// NewPublicWorkerBootHandler checks the parent's existing identity without importing child resource fields.
func NewPublicWorkerBootHandler(shr *share, bootErr *error) *UnifiedBootHandler {
	expectedToken := shr.token
	return NewBootHandler(&BootHandlerConfig{OperationType: "public worker", validate: func(msg subordinate.Message) error {
		boot, err := dd.New[publicWorkerBoot](msg)
		if err != nil {
			return fmt.Errorf("invalid public worker boot: %w", err)
		}
		if strings.TrimSpace(expectedToken) == "" || strings.TrimSpace(boot.Token) == "" || boot.Token != expectedToken {
			return fmt.Errorf("public worker boot token '%v' does not match expected token '%v'", boot.Token, expectedToken)
		}
		return nil
	}}, bootErr)
}
