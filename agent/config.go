package agent

type AgentConfig struct {
	ConsoleAddress   string
	ConsoleStartPort uint16
	ConsoleEndPort   uint16
}

func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		ConsoleAddress:   "127.0.0.1",
		ConsoleStartPort: 8080,
		ConsoleEndPort:   8181,
	}
}
