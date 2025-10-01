package metrics

import (
	"github.com/michaelquigley/df/dd"
)

type Config struct {
	Influx *InfluxConfig
	Agent  *AgentConfig
}

type AgentConfig struct {
	Source dd.Dynamic
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string `dd:"+secret"`
}
