package canary

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type DisablerOptions struct {
	Environments chan *sdk.Environment
	MinDwell     time.Duration
	MaxDwell     time.Duration
	MinPacing    time.Duration
	MaxPacing    time.Duration
}

type Disabler struct {
	Id   uint
	Done chan struct{}
	opt  *DisablerOptions
	root env_core.Root
}

func NewDisabler(id uint, opt *DisablerOptions, root env_core.Root) *Disabler {
	return &Disabler{
		Id:   id,
		Done: make(chan struct{}),
		opt:  opt,
		root: root,
	}
}

func (d *Disabler) Run() {
	defer logrus.Infof("#%d stopping", d.Id)
	defer close(d.Done)
	d.dwell()
	d.iterate()
}

func (d *Disabler) dwell() {
	dwell := d.opt.MinDwell.Milliseconds()
	dwelta := d.opt.MaxDwell.Milliseconds() - d.opt.MinDwell.Milliseconds()
	if dwelta > 0 {
		dwell = int64(rand.Intn(int(dwelta)) + int(d.opt.MinDwell.Milliseconds()))
	}
	time.Sleep(time.Duration(dwell) * time.Millisecond)
}

func (d *Disabler) iterate() {
	for {
		select {
		case env, ok := <-d.opt.Environments:
			if !ok {
				return
			}
			if err := sdk.DisableEnvironment(env, d.root); err == nil {
				logrus.Infof("#%d disabled environment '%v'", d.Id, env.ZitiIdentity)
			} else {
				logrus.Errorf("error disabling canary (#%d) environment '%v': %v", d.Id, env.ZitiIdentity, err)
			}
		}

		pacingMs := d.opt.MaxPacing.Milliseconds()
		pacingDelta := d.opt.MaxPacing.Milliseconds() - d.opt.MinPacing.Milliseconds()
		if pacingDelta > 0 {
			pacingMs = (rand.Int63() % pacingDelta) + d.opt.MinPacing.Milliseconds()
			time.Sleep(time.Duration(pacingMs) * time.Millisecond)
		}
	}
}
