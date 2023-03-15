package metrics2

type AgentConfig struct {
	Source interface{}
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string `cf:"+secret"`
}
