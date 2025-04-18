package canary

import (
	"fmt"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type EnablerOptions struct {
	Iterations uint
	MinDwell   time.Duration
	MaxDwell   time.Duration
	MinPacing  time.Duration
	MaxPacing  time.Duration
}

type Enabler struct {
	Id           uint
	opt          *EnablerOptions
	root         env_core.Root
	Environments chan *sdk.Environment
}

func NewEnabler(id uint, opt *EnablerOptions, root env_core.Root) *Enabler {
	return &Enabler{
		Id:           id,
		opt:          opt,
		root:         root,
		Environments: make(chan *sdk.Environment, opt.Iterations),
	}
}

func (e *Enabler) Run() {
	defer close(e.Environments)
	defer logrus.Infof("#%d stopping", e.Id)
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
	for i := uint(0); i < e.opt.Iterations; i++ {
		env, err := sdk.EnableEnvironment(e.root, &sdk.EnableRequest{
			Host:        fmt.Sprintf("canary_%d_%d", e.Id, i),
			Description: "canary.Enabler",
		})
		if err == nil {
			e.Environments <- env
			logrus.Infof("#%d enabled environment '%v'", e.Id, env.ZitiIdentity)
		} else {
			logrus.Errorf("error creating canary (#%d) environment: %v", e.Id, err)
		}

		pacingMs := e.opt.MaxPacing.Milliseconds()
		pacingDelta := e.opt.MaxPacing.Milliseconds() - e.opt.MinPacing.Milliseconds()
		if pacingDelta > 0 {
			pacingMs = (rand.Int63() % pacingDelta) + e.opt.MinPacing.Milliseconds()
			time.Sleep(time.Duration(pacingMs) * time.Millisecond)
		}
	}
}
