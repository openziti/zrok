package metrics

type Config struct {
	Influx *InfluxConfig
	Agent  *AgentConfig
}

type AgentConfig struct {
	Source interface{}
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string `cf:"+secret"`
}
