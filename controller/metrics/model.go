package metrics

type Source interface {
	Start(chan map[string]interface{}) (chan struct{}, error)
	Stop()
}

type Ingester interface {
	Ingest(msg map[string]interface{}) error
}
