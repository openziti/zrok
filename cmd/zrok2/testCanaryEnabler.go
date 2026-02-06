package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/canary"
	"github.com/openziti/zrok/v2/environment"
	"github.com/spf13/cobra"
)

func init() {
	testCanaryCmd.AddCommand(newTestCanaryEnabler().cmd)
}

type testCanaryEnabler struct {
	cmd          *cobra.Command
	enablers     uint
	iterations   uint
	preDelay     time.Duration
	minPreDelay  time.Duration
	maxPreDelay  time.Duration
	dwell        time.Duration
	minDwell     time.Duration
	maxDwell     time.Duration
	pacing       time.Duration
	minPacing    time.Duration
	maxPacing    time.Duration
	skipDisable  bool
	canaryConfig string
}

func newTestCanaryEnabler() *testCanaryEnabler {
	cmd := &cobra.Command{
		Use:   "enabler",
		Short: "Enable a canary enabling environments",
		Args:  cobra.NoArgs,
	}
	command := &testCanaryEnabler{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().UintVarP(&command.enablers, "enablers", "e", 1, "Number of concurrent enablers to start")
	cmd.Flags().UintVarP(&command.iterations, "iterations", "i", 1, "Number of iterations")
	cmd.Flags().DurationVar(&command.dwell, "dwell", 0, "Fixed dwell time")
	cmd.Flags().DurationVar(&command.minDwell, "min-dwell", 0, "Minimum dwell time")
	cmd.Flags().DurationVar(&command.maxDwell, "max-dwell", 0, "Maximum dwell time")
	cmd.Flags().DurationVar(&command.pacing, "pacing", 0, "Fixed pacing time")
	cmd.Flags().DurationVar(&command.minPacing, "min-pacing", 0, "Minimum pacing time")
	cmd.Flags().DurationVar(&command.maxPacing, "max-pacing", 0, "Maximum pacing time")
	cmd.Flags().BoolVar(&command.skipDisable, "skip-disable", false, "Disable (clean up) enabled environments")
	cmd.Flags().StringVar(&command.canaryConfig, "canary-config", "", "Path to canary configuration file")
	return command
}

func (cmd *testCanaryEnabler) run(_ *cobra.Command, _ []string) {
	if err := canary.AcknowledgeDangerousCanary(); err != nil {
		dl.Fatal(err)
	}

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
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

	var enablers []*canary.Enabler
	for i := uint(0); i < cmd.enablers; i++ {
		preDelay := cmd.maxPreDelay.Milliseconds()
		preDelayDelta := cmd.maxPreDelay.Milliseconds() - cmd.minPreDelay.Milliseconds()
		if preDelayDelta > 0 {
			preDelay = int64(rand.Intn(int(preDelayDelta))) + cmd.minPreDelay.Milliseconds()
		}
		time.Sleep(time.Duration(preDelay) * time.Millisecond)

		enablerOpts := &canary.EnablerOptions{
			Iterations: cmd.iterations,
			MinDwell:   cmd.minDwell,
			MaxDwell:   cmd.maxDwell,
			MinPacing:  cmd.minPacing,
			MaxPacing:  cmd.maxPacing,
		}
		if cmd.pacing > 0 {
			enablerOpts.MinDwell = cmd.dwell
			enablerOpts.MaxDwell = cmd.dwell
		}
		if cmd.pacing > 0 {
			enablerOpts.MinPacing = cmd.pacing
			enablerOpts.MaxPacing = cmd.pacing
		}
		if sns != nil {
			enablerOpts.SnapshotQueue = sns.InputQueue
		}
		enabler := canary.NewEnabler(i, enablerOpts, root)
		enablers = append(enablers, enabler)
		go enabler.Run()
	}

	if !cmd.skipDisable {
		var disablers []*canary.Disabler
		for i := uint(0); i < cmd.enablers; i++ {
			disablerOpts := &canary.DisablerOptions{
				Environments: enablers[i].Environments,
			}
			if sns != nil {
				disablerOpts.SnapshotQueue = sns.InputQueue
			}
			disabler := canary.NewDisabler(i, disablerOpts, root)
			disablers = append(disablers, disabler)
			go disabler.Run()
		}
		for _, disabler := range disablers {
			dl.Infof("waiting for disabler #%d", disabler.Id)
			<-disabler.Done
		}

	} else {
		for _, enabler := range enablers {
		enablerLoop:
			for {
				select {
				case env, ok := <-enabler.Environments:
					if !ok {
						break enablerLoop
					}
					dl.Infof("enabler #%d: %v", enabler.Id, env.ZitiIdentity)
				}
			}
		}
	}

	for _, enabler := range enablers {
		<-enabler.Done
	}

	if sns != nil {
		snsCancel()
		<-sns.Closed
	}

	dl.Info("complete")
}
