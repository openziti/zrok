package main

import (
	"context"
	"github.com/openziti/zrok/canary"
	"github.com/openziti/zrok/environment"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	testCanaryCmd.AddCommand(newTestCanaryPublicProxy().cmd)
}

type testCanaryPublicProxy struct {
	cmd               *cobra.Command
	loopers           uint
	iterations        uint
	statusInterval    uint
	timeout           time.Duration
	payload           uint64
	minPayload        uint64
	maxPayload        uint64
	preDelay          time.Duration
	minPreDelay       time.Duration
	maxPreDelay       time.Duration
	dwell             time.Duration
	minDwell          time.Duration
	maxDwell          time.Duration
	pacing            time.Duration
	minPacing         time.Duration
	maxPacing         time.Duration
	batchSize         uint
	batchPacing       time.Duration
	minBatchPacing    time.Duration
	maxBatchPacing    time.Duration
	frontendSelection string
	canaryConfig      string
}

func newTestCanaryPublicProxy() *testCanaryPublicProxy {
	cmd := &cobra.Command{
		Use:   "public-proxy",
		Short: "Run a public `proxy` looper canary",
		Args:  cobra.NoArgs,
	}
	command := &testCanaryPublicProxy{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().UintVarP(&command.loopers, "loopers", "l", 1, "Number of concurrent loopers to start")
	cmd.Flags().UintVarP(&command.iterations, "iterations", "i", 1, "Number of iterations")
	cmd.Flags().UintVarP(&command.statusInterval, "status-interval", "S", 100, "Show status every # iterations")
	cmd.Flags().DurationVarP(&command.timeout, "timeout", "T", 30*time.Second, "Timeout when sending HTTP requests")
	cmd.Flags().Uint64Var(&command.payload, "payload", 0, "Fixed payload size")
	cmd.Flags().Uint64Var(&command.minPayload, "min-payload", 64, "Minimum payload size in bytes")
	cmd.Flags().Uint64Var(&command.maxPayload, "max-payload", 10240, "Maximum payload size in bytes")
	cmd.Flags().DurationVar(&command.dwell, "dwell", 1*time.Second, "Fixed dwell time")
	cmd.Flags().DurationVar(&command.minDwell, "min-dwell", 1*time.Second, "Minimum dwell time")
	cmd.Flags().DurationVar(&command.maxDwell, "max-dwell", 1*time.Second, "Maximum dwell time")
	cmd.Flags().DurationVar(&command.pacing, "pacing", 0, "Fixed pacing time")
	cmd.Flags().DurationVar(&command.minPacing, "min-pacing", 0, "Minimum pacing time")
	cmd.Flags().DurationVar(&command.maxPacing, "max-pacing", 0, "Maximum pacing time")
	cmd.Flags().UintVar(&command.batchSize, "batch-size", 0, "Iterate in batches of this size")
	cmd.Flags().DurationVar(&command.batchPacing, "batch-pacing", 0, "Fixed batch pacing time")
	cmd.Flags().DurationVar(&command.minBatchPacing, "min-batch-pacing", 0, "Minimum batch pacing time")
	cmd.Flags().DurationVar(&command.maxBatchPacing, "max-batch-pacing", 0, "Maximum batch pacing time")
	cmd.Flags().StringVar(&command.frontendSelection, "frontend-selection", "public", "Select frontend selection")
	cmd.Flags().StringVar(&command.canaryConfig, "canary-config", "", "Path to canary configuration file")
	return command
}

func (cmd *testCanaryPublicProxy) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	if !root.IsEnabled() {
		logrus.Fatal("unable to load environment; did you 'zrok enable'?")
	}

	var sc *canary.SnapshotStreamer
	var scCtx context.Context
	var scCancel context.CancelFunc
	if cmd.canaryConfig != "" {
		cfg, err := canary.LoadConfig(cmd.canaryConfig)
		if err != nil {
			panic(err)
		}
		scCtx, scCancel = context.WithCancel(context.Background())
		sc, err = canary.NewSnapshotStreamer(scCtx, cfg)
		if err != nil {
			panic(err)
		}
		go sc.Run()
	}

	var loopers []*canary.PublicHttpLooper
	for i := uint(0); i < cmd.loopers; i++ {
		preDelay := cmd.maxPreDelay.Milliseconds()
		preDelayDelta := cmd.maxPreDelay.Milliseconds() - cmd.minPreDelay.Milliseconds()
		if preDelayDelta > 0 {
			preDelay = int64(rand.Intn(int(preDelayDelta))) + cmd.minPreDelay.Milliseconds()
			time.Sleep(time.Duration(preDelay) * time.Millisecond)
		}

		looperOpts := &canary.LooperOptions{
			Iterations:     cmd.iterations,
			StatusInterval: cmd.statusInterval,
			Timeout:        cmd.timeout,
			BatchSize:      cmd.batchSize,
		}
		if cmd.payload > 0 {
			looperOpts.MinPayload = cmd.payload
			looperOpts.MaxPayload = cmd.payload
		} else {
			looperOpts.MinPayload = cmd.minPayload
			looperOpts.MaxPayload = cmd.maxPayload
		}
		if cmd.dwell > 0 {
			looperOpts.MinDwell = cmd.dwell
			looperOpts.MaxDwell = cmd.dwell
		} else {
			looperOpts.MinDwell = cmd.minDwell
			looperOpts.MaxDwell = cmd.maxDwell
		}
		if cmd.pacing > 0 {
			looperOpts.MinPacing = cmd.pacing
			looperOpts.MaxPacing = cmd.pacing
		} else {
			looperOpts.MinPacing = cmd.minPacing
			looperOpts.MaxPacing = cmd.maxPacing
		}
		if cmd.batchPacing > 0 {
			looperOpts.MinBatchPacing = cmd.batchPacing
			looperOpts.MaxBatchPacing = cmd.batchPacing
		} else {
			looperOpts.MinBatchPacing = cmd.minBatchPacing
			looperOpts.MaxBatchPacing = cmd.maxBatchPacing
		}
		if sc != nil {
			looperOpts.SnapshotQueue = sc.InputQueue
		}
		looper := canary.NewPublicHttpLooper(i, cmd.frontendSelection, looperOpts, root)
		loopers = append(loopers, looper)
		go looper.Run()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		for _, looper := range loopers {
			looper.Abort()
		}
	}()

	for _, l := range loopers {
		<-l.Done()
	}

	if sc != nil {
		scCancel()
		<-sc.Closed
	}

	results := make([]*canary.LooperResults, 0)
	for i := uint(0); i < cmd.loopers; i++ {
		results = append(results, loopers[i].Results())
	}
	canary.ReportLooperResults(results)

	os.Exit(0)
}
