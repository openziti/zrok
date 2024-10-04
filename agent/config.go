package agent

type AgentConfig struct {
	ConsoleEndpoint string
}

func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		ConsoleEndpoint: "127.0.0.1:8888",
	}
}
