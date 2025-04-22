package canary

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
	"slices"
	"sort"
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

func (s *Snapshot) String() string {
	if s.Ok {
		return fmt.Sprintf("[%v, %d, %d] (ok) %v, %v", s.Operation, s.Instance, s.Iteration, s.Completed.Sub(s.Started), util.BytesToSize(int64(s.Size)))
	} else {
		return fmt.Sprintf("[%v, %d, %d] (err) %v, %v, (%v)", s.Operation, s.Instance, s.Iteration, s.Completed.Sub(s.Started), util.BytesToSize(int64(s.Size)), s.Error)
	}
}

type SnapshotCollector struct {
	InputQueue chan *Snapshot
	Closed     chan struct{}
	ctx        context.Context
	cfg        *Config
	snapshots  map[string][]*Snapshot
}

func NewSnapshotCollector(ctx context.Context, cfg *Config) *SnapshotCollector {
	return &SnapshotCollector{
		InputQueue: make(chan *Snapshot),
		Closed:     make(chan struct{}),
		ctx:        ctx,
		cfg:        cfg,
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

func (sc *SnapshotCollector) Store() error {
	idb := influxdb2.NewClient(sc.cfg.Influx.Url, sc.cfg.Influx.Token)
	writeApi := idb.WriteAPIBlocking(sc.cfg.Influx.Org, sc.cfg.Influx.Bucket)
	for key, arr := range sc.snapshots {
		for _, snapshot := range arr {
			tags := map[string]string{
				"instance":  fmt.Sprintf("%d", snapshot.Instance),
				"iteration": fmt.Sprintf("%d", snapshot.Iteration),
				"ok":        fmt.Sprintf("%t", snapshot.Ok),
			}
			if snapshot.Error != nil {
				tags["error"] = snapshot.Error.Error()
			}
			pt := influxdb2.NewPoint(snapshot.Operation, tags, map[string]interface{}{
				"duration": snapshot.Completed.Sub(snapshot.Started).Milliseconds(),
				"size":     snapshot.Size,
			}, snapshot.Started)
			if err := writeApi.WritePoint(context.Background(), pt); err != nil {
				return err
			}
		}
		logrus.Infof("wrote '%v' points for '%v'", len(arr), key)
	}
	idb.Close()
	logrus.Infof("complete")
	return nil
}
