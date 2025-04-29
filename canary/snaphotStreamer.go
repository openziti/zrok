package canary

import (
	"context"
	"errors"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/sirupsen/logrus"
)

type SnapshotStreamer struct {
	InputQueue  chan *Snapshot
	Closed      chan struct{}
	ctx         context.Context
	cfg         *Config
	ifxClient   influxdb2.Client
	ifxWriteApi api.WriteAPIBlocking
}

func NewSnapshotStreamer(ctx context.Context, cfg *Config) (*SnapshotStreamer, error) {
	out := &SnapshotStreamer{
		InputQueue: make(chan *Snapshot),
		Closed:     make(chan struct{}),
		ctx:        ctx,
		cfg:        cfg,
	}
	if cfg.Influx != nil {
		out.ifxClient = influxdb2.NewClient(cfg.Influx.Url, cfg.Influx.Token)
		out.ifxWriteApi = out.ifxClient.WriteAPIBlocking(cfg.Influx.Org, cfg.Influx.Bucket)
	} else {
		return nil, errors.New("missing influx configuration")
	}
	return out, nil
}

func (ss *SnapshotStreamer) Run() {
	defer close(ss.Closed)
	defer ss.ifxClient.Close()
	defer logrus.Info("stoping")
	logrus.Info("starting")

	for {
		select {
		case <-ss.ctx.Done():
			return

		case snapshot := <-ss.InputQueue:
			if err := ss.store(snapshot); err != nil {
				logrus.Errorf("error storing snapshot: %v", err)
			}
		}
	}
}

func (ss *SnapshotStreamer) store(snapshot *Snapshot) error {
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
	if err := ss.ifxWriteApi.WritePoint(context.Background(), pt); err != nil {
		return err
	}
	return nil
}
