package canary

import (
	"fmt"
	"github.com/openziti/zrok/util"
	"time"
)

type Snapshot struct {
	Operation string
	Instance  uint
	Iteration uint64
	Started   time.Time
	Completed time.Time
	Ok        bool
	Error     error
	Size      uint64
}

func NewSnapshot(operation string, instance uint, iteration uint64) *Snapshot {
	return &Snapshot{Operation: operation, Instance: instance, Iteration: iteration, Started: time.Now()}
}

func (s *Snapshot) Complete() *Snapshot {
	s.Completed = time.Now()
	return s
}

func (s *Snapshot) Success() *Snapshot {
	s.Ok = true
	return s
}

func (s *Snapshot) Failure(err error) *Snapshot {
	s.Ok = false
	s.Error = err
	return s
}

func (s *Snapshot) Send(queue chan *Snapshot) {
	if queue != nil {
		queue <- s
	}
}

func (s *Snapshot) String() string {
	if s.Ok {
		return fmt.Sprintf("[%v, %d, %d] (ok) %v, %v", s.Operation, s.Instance, s.Iteration, s.Completed.Sub(s.Started), util.BytesToSize(int64(s.Size)))
	} else {
		return fmt.Sprintf("[%v, %d, %d] (err) %v, %v, (%v)", s.Operation, s.Instance, s.Iteration, s.Completed.Sub(s.Started), util.BytesToSize(int64(s.Size)), s.Error)
	}
}
