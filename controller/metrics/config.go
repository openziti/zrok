package metrics

type Config struct {
	Influx     *InfluxConfig
	Strategies *StrategiesConfig
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string `cf:"+secret"`
}

type StrategiesConfig struct {
	Source interface{}
}
