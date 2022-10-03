package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newRun().cmd)
}

type run struct {
	cmd     *cobra.Command
	loopers int
}

func newRun() *run {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start a loop agent",
		Args:  cobra.ExactArgs(0),
	}
	r := &run{cmd: cmd}
	cmd.Run = r.run
	cmd.Flags().IntVarP(&r.loopers, "loopers", "l", 1, "Number of current loopers to start")
	return r
}

func (r *run) run(_ *cobra.Command, _ []string) {
	var loopers []*looper
	for i := 0; i < r.loopers; i++ {
		l := newLooper(i)
		loopers = append(loopers, l)
		go l.run()
	}
	for _, l := range loopers {
		<-l.done
	}
}

type looper struct {
	id   int
	done chan struct{}
}

func newLooper(id int) *looper {
	return &looper{
		id:   id,
		done: make(chan struct{}),
	}
}

func (l *looper) run() {
	logrus.Infof("starting #%d", l.id)
	defer close(l.done)
	defer logrus.Infof("stopping #%d", l.id)
}
