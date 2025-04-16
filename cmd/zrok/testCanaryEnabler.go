package main

import (
	"github.com/openziti/zrok/canary"
	"github.com/openziti/zrok/environment"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"math/rand"
	"time"
)

func init() {
	testCanaryCmd.AddCommand(newTestCanaryEnabler().cmd)
}

type testCanaryEnabler struct {
	cmd         *cobra.Command
	enablers    uint
	iterations  uint
	minPreDelay time.Duration
	maxPreDelay time.Duration
	minDwell    time.Duration
	maxDwell    time.Duration
	minPacing   time.Duration
	maxPacing   time.Duration
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
	cmd.Flags().DurationVar(&command.minDwell, "min-dwell", 1*time.Second, "Minimum dwell time")
	cmd.Flags().DurationVar(&command.maxDwell, "max-dwell", 1*time.Second, "Maximum dwell time")
	cmd.Flags().DurationVar(&command.minPacing, "min-pacing", 0, "Minimum pacing time")
	cmd.Flags().DurationVar(&command.maxPacing, "max-pacing", 0, "Maximum pacing time")
	return command
}

func (cmd *testCanaryEnabler) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	var enablers []*canary.Enabler
	for i := uint(0); i < cmd.enablers; i++ {
		preDelay := cmd.maxPreDelay.Milliseconds()
		preDelayDelta := cmd.maxPreDelay.Milliseconds() - cmd.minPreDelay.Milliseconds()
		if preDelayDelta > 0 {
			preDelay = int64(rand.Intn(int(preDelayDelta))) + cmd.minPreDelay.Milliseconds()
			time.Sleep(time.Duration(preDelay) * time.Millisecond)
		}

		enablerOpts := &canary.EnablerOptions{
			Iterations: cmd.iterations,
			MinDwell:   cmd.minDwell,
			MaxDwell:   cmd.maxDwell,
			MinPacing:  cmd.minPacing,
			MaxPacing:  cmd.maxPacing,
		}
		enabler := canary.NewEnabler(i, enablerOpts, root)
		enablers = append(enablers, enabler)
		go enabler.Run()
	}

	for _, enabler := range enablers {
	enablerLoop:
		for {
			select {
			case env, ok := <-enabler.Environments:
				if !ok {
					break enablerLoop
				}
				logrus.Infof("enabler #%d: %v", enabler.Id, env.ZitiIdentity)
			}
		}
	}
}
