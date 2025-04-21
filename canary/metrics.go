package canary

import (
	"context"
	"github.com/sirupsen/logrus"
	"slices"
	"sort"
	"time"
)

type Snapshot struct {
	Stamp     time.Time
	Operation string
	Instance  uint
	Started   time.Time
	Completed time.Time
	Size      uint64
}

type SnapshotCollector struct {
	InputQueue chan *Snapshot
	Closed     chan struct{}
	ctx        context.Context
	snapshots  map[string][]*Snapshot
}

func NewSnapshotCollector(ctx context.Context) *SnapshotCollector {
	return &SnapshotCollector{
		InputQueue: make(chan *Snapshot),
		Closed:     make(chan struct{}),
		ctx:        ctx,
		snapshots:  make(map[string][]*Snapshot),
	}
}

func (sc *SnapshotCollector) Run() {
	defer close(sc.Closed)
	defer logrus.Info("stopping")
	logrus.Info("starting")
	for {
		select {
		case <-sc.ctx.Done():
			return

		case snapshot := <-sc.InputQueue:
			var snapshots []*Snapshot
			if v, ok := sc.snapshots[snapshot.Operation]; ok {
				snapshots = v
			}
			i := sort.Search(len(snapshots), func(i int) bool { return snapshots[i].Completed.After(snapshot.Started) })
			snapshots = slices.Insert(snapshots, i, snapshot)
			sc.snapshots[snapshot.Operation] = snapshots
		}
	}
}
