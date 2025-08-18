package metrics

import "github.com/michaelquigley/df"

type Config struct {
	Influx *InfluxConfig
	Agent  *AgentConfig
}

type AgentConfig struct {
	Source df.Dynamic
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string `df:",secret"`
}
