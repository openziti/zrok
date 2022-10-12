package controller

type MetricsConfig struct {
	Influx *InfluxConfig
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string
}
