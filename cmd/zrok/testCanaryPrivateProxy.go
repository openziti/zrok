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
	testCanaryCmd.AddCommand(newTestCanaryPrivateProxy().cmd)
}

type testCanaryPrivateProxy struct {
	cmd            *cobra.Command
	loopers        uint
	iterations     uint
	statusInterval uint
	timeout        time.Duration
	payload        uint64
	minPayload     uint64
	maxPayload     uint64
	preDelay       time.Duration
	minPreDelay    time.Duration
	maxPreDelay    time.Duration
	dwell          time.Duration
	minDwell       time.Duration
	maxDwell       time.Duration
	pacing         time.Duration
	minPacing      time.Duration
	maxPacing      time.Duration
	batchSize      uint
	batchPacing    time.Duration
	minBatchPacing time.Duration
	maxBatchPacing time.Duration
	targetName     string
	bindAddress    string
	canaryConfig   string
}

func newTestCanaryPrivateProxy() *testCanaryPrivateProxy {
	cmd := &cobra.Command{
		Use:   "private-proxy",
		Short: "Run a private `proxy` looper canary",
		Args:  cobra.NoArgs,
	}
	command := &testCanaryPrivateProxy{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().UintVarP(&command.loopers, "loopers", "l", 1, "Number of concurrent loopers to start")
	cmd.Flags().UintVarP(&command.iterations, "iterations", "i", 1, "Number of iterations")
	cmd.Flags().UintVarP(&command.statusInterval, "status-interval", "S", 100, "Show status every # iterations")
	cmd.Flags().DurationVarP(&command.timeout, "timeout", "T", 30*time.Second, "Timeout when sending HTTP requests")
	cmd.Flags().Uint64Var(&command.payload, "payload", 0, "Fixed payload size in bytes")
	cmd.Flags().Uint64Var(&command.minPayload, "min-payload", 64, "Minimum payload size in bytes")
	cmd.Flags().Uint64Var(&command.maxPayload, "max-payload", 10240, "Maximum payload size in bytes")
	cmd.Flags().DurationVar(&command.preDelay, "pre-delay", 0, "Fixed pre-delay before creating the next looper")
	cmd.Flags().DurationVar(&command.minPreDelay, "min-pre-delay", 0, "Minimum pre-delay before creating the next looper")
	cmd.Flags().DurationVar(&command.maxPreDelay, "max-pre-delay", 0, "Maximum pre-delay before creating the next looper")
	cmd.Flags().DurationVar(&command.dwell, "dwell", 0, "Fixed dwell time")
	cmd.Flags().DurationVar(&command.minDwell, "min-dwell", 1*time.Second, "Minimum dwell time")
	cmd.Flags().DurationVar(&command.maxDwell, "max-dwell", 1*time.Second, "Maximum dwell time")
	cmd.Flags().DurationVar(&command.pacing, "pacing", 0, "Fixed pacing time")
	cmd.Flags().DurationVar(&command.minPacing, "min-pacing", 0, "Minimum pacing time")
	cmd.Flags().DurationVar(&command.maxPacing, "max-pacing", 0, "Maximum pacing time")
	cmd.Flags().UintVar(&command.batchSize, "batch-size", 0, "Iterate in batches of this size")
	cmd.Flags().DurationVar(&command.batchPacing, "batch-pacing", 0, "Fixed batch pacing time")
	cmd.Flags().DurationVar(&command.minBatchPacing, "min-batch-pacing", 0, "Minimum batch pacing time")
	cmd.Flags().DurationVar(&command.maxBatchPacing, "max-batch-pacing", 0, "Maximum batch pacing time")
	cmd.Flags().StringVar(&command.targetName, "target-name", "", "Metadata describing the virtual target")
	cmd.Flags().StringVar(&command.bindAddress, "bind-address", "", "Metadata describing the virtual bind address")
	cmd.Flags().StringVar(&command.canaryConfig, "canary-config", "", "Path to canary configuration file")
	return command
}

func (cmd *testCanaryPrivateProxy) run(_ *cobra.Command, _ []string) {
	if err := canary.AcknowledgeDangerousCanary(); err != nil {
		logrus.Fatal(err)
	}

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	if !root.IsEnabled() {
		logrus.Fatal("unable to load environment; did you 'zrok enable'?")
	}

	var sns *canary.SnapshotStreamer
	var snsCtx context.Context
	var snsCancel context.CancelFunc
	if cmd.canaryConfig != "" {
		cfg, err := canary.LoadConfig(cmd.canaryConfig)
		if err != nil {
			panic(err)
		}
		snsCtx, snsCancel = context.WithCancel(context.Background())
		sns, err = canary.NewSnapshotStreamer(snsCtx, cfg)
		if err != nil {
			panic(err)
		}
		go sns.Run()
	}

	var loopers []*canary.PrivateHttpLooper
	for i := uint(0); i < cmd.loopers; i++ {
		var preDelay int64
		if cmd.preDelay > 0 {
			preDelay = cmd.preDelay.Milliseconds()
		} else {
			preDelay = cmd.maxPreDelay.Milliseconds()
			preDelayDelta := cmd.maxPreDelay.Milliseconds() - cmd.minPreDelay.Milliseconds()
			if preDelayDelta > 0 {
				preDelay = int64(rand.Intn(int(preDelayDelta))) + cmd.minPreDelay.Milliseconds()
			}
		}
		time.Sleep(time.Duration(preDelay) * time.Millisecond)

		looperOpts := &canary.LooperOptions{
			Iterations:     cmd.iterations,
			StatusInterval: cmd.statusInterval,
			Timeout:        cmd.timeout,
			MinPayload:     cmd.minPayload,
			MaxPayload:     cmd.maxPayload,
			MinDwell:       cmd.minDwell,
			MaxDwell:       cmd.maxDwell,
			MinPacing:      cmd.minPacing,
			MaxPacing:      cmd.maxPacing,
			BatchSize:      cmd.batchSize,
			MinBatchPacing: cmd.minBatchPacing,
			MaxBatchPacing: cmd.maxBatchPacing,
			TargetName:     cmd.targetName,
			BindAddress:    cmd.bindAddress,
		}
		if cmd.payload > 0 {
			looperOpts.MinPayload = cmd.payload
			looperOpts.MaxPayload = cmd.payload
		}
		if cmd.dwell > 0 {
			looperOpts.MinDwell = cmd.dwell
			looperOpts.MaxDwell = cmd.dwell
		}
		if cmd.pacing > 0 {
			looperOpts.MinPacing = cmd.pacing
			looperOpts.MaxPacing = cmd.pacing
		}
		if cmd.batchPacing > 0 {
			looperOpts.MinBatchPacing = cmd.batchPacing
			looperOpts.MaxBatchPacing = cmd.batchPacing
		}
		if sns != nil {
			looperOpts.SnapshotQueue = sns.InputQueue
		}
		looper := canary.NewPrivateHttpLooper(i, looperOpts, root)
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

	if sns != nil {
		snsCancel()
		<-sns.Closed
	}

	results := make([]*canary.LooperResults, 0)
	for i := uint(0); i < cmd.loopers; i++ {
		results = append(results, loopers[i].Results())
	}
	canary.ReportLooperResults(results)

	os.Exit(0)
}
