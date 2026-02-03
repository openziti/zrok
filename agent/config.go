package agent

import "time"

type AgentConfig struct {
	ConsoleAddress     string        `json:"console_address"`
	ConsoleStartPort   uint16        `json:"console_start_port"`
	ConsoleEndPort     uint16        `json:"console_end_port"`
	ConsoleEnabled     bool          `json:"console_enabled"`
	RequireRemoting    bool          `json:"require_remoting"`
	MaxRetries         int           `json:"max_retries,omitempty"`
	RetryInitialDelay  time.Duration `json:"retry_initial_delay,omitempty"`
	RetryMaxDelay      time.Duration `json:"retry_max_delay,omitempty"`
	RetryCheckInterval time.Duration `json:"retry_check_interval,omitempty"`
}

func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		ConsoleAddress:     "127.0.0.1",
		ConsoleStartPort:   8080,
		ConsoleEndPort:     8181,
		ConsoleEnabled:     true,
		RequireRemoting:    false,
		MaxRetries:         -1, // unlimited
		RetryInitialDelay:  30 * time.Second,
		RetryMaxDelay:      8 * time.Minute,
		RetryCheckInterval: 30 * time.Second,
	}
}
