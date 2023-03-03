package metrics

type Source interface {
	Start() (chan struct{}, error)
	Stop()
}

type Ingester interface {
	Ingest(msg map[string]interface{})
}
