package agent

type Daemon struct {
	shares   map[string]*share
	accesses map[string]*access
}

func NewDaemon() *Daemon {
	return &Daemon{
		shares:   make(map[string]*share),
		accesses: make(map[string]*access),
	}
}
