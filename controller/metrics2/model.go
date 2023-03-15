package metrics2

type ZitiEventJson string

type ZitiEventJsonSource interface {
	Start(chan ZitiEventJson) (join chan struct{}, err error)
	Stop()
}

type ZitiEventJsonSink interface {
	Handle(event ZitiEventJson) error
}
