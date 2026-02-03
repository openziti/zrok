package canary

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

type EnablerOptions struct {
	Iterations    uint
	MinDwell      time.Duration
	MaxDwell      time.Duration
	MinPacing     time.Duration
	MaxPacing     time.Duration
	SnapshotQueue chan *Snapshot
}

type Enabler struct {
	Id           uint
	Done         chan struct{}
	opt          *EnablerOptions
	root         env_core.Root
	Environments chan *sdk.Environment
}

func NewEnabler(id uint, opt *EnablerOptions, root env_core.Root) *Enabler {
	return &Enabler{
		Id:           id,
		Done:         make(chan struct{}),
		opt:          opt,
		root:         root,
		Environments: make(chan *sdk.Environment, opt.Iterations),
	}
}

func (e *Enabler) Run() {
	defer close(e.Environments)
	defer close(e.Done)
	defer dl.Infof("#%d stopping", e.Id)
	e.dwell()
	e.iterate()
}

func (e *Enabler) dwell() {
	dwell := e.opt.MinDwell.Milliseconds()
	dwelta := e.opt.MaxDwell.Milliseconds() - e.opt.MinDwell.Milliseconds()
	if dwelta > 0 {
		dwell = int64(rand.Intn(int(dwelta)) + int(e.opt.MinDwell.Milliseconds()))
	}
	time.Sleep(time.Duration(dwell) * time.Millisecond)
}

func (e *Enabler) iterate() {
	defer dl.Info("done")
	for i := uint(0); i < e.opt.Iterations; i++ {
		snapshot := NewSnapshot("enable", e.Id, uint64(i))

		env, err := sdk.EnableEnvironment(e.root, &sdk.EnableRequest{
			Host:        fmt.Sprintf("canary_%d_%d", e.Id, i),
			Description: "canary.Enabler",
		})
		if err == nil {
			snapshot.Complete().Success()
			e.Environments <- env
			dl.Infof("#%d enabled environment '%v'", e.Id, env.ZitiIdentity)

		} else {
			snapshot.Complete().Failure(err)

			dl.Errorf("error creating canary (#%d) environment: %v", e.Id, err)
		}

		snapshot.Send(e.opt.SnapshotQueue)

		pacingMs := e.opt.MaxPacing.Milliseconds()
		pacingDelta := e.opt.MaxPacing.Milliseconds() - e.opt.MinPacing.Milliseconds()
		if pacingDelta > 0 {
			pacingMs = (rand.Int63() % pacingDelta) + e.opt.MinPacing.Milliseconds()
		}
		time.Sleep(time.Duration(pacingMs) * time.Millisecond)
	}
}
