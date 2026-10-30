//go:build agent_smoke && unix

// agent-smoke drives the named public share shutdown and restart scenario against a live zrok service, using a
// separately built zrok2 binary. it is built only with '-tags agent_smoke'. see docs/current/agent-smoke.md.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/michaelquigley/df/dl"
)

const usage = "usage: agent-smoke {run|cleanup} --binary <zrok2> --state-dir <dir> [--api-endpoint <url>] [--namespace <token>]"

func main() {
	os.Exit(entry(os.Args[1:], os.Stderr))
}

func entry(args []string, stderr io.Writer) int {
	// workflow logs are read by people, so plain lines even without a terminal
	logOptions := dl.DefaultOptions()
	logOptions.UseJSON = false
	dl.Init(logOptions)

	mode, cfg, err := parseArgs(args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n%v\n", err, usage)
		return 2
	}

	// the first interrupt cancels the scenario; a second abandons cleanup, leaving the workflow's fallback step
	interrupts := make(chan os.Signal, 2)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupts)
	first, cancelFirst := context.WithCancel(context.Background())
	defer cancelFirst()
	second, cancelSecond := context.WithCancel(context.Background())
	defer cancelSecond()
	go func() {
		count := 0
		for {
			select {
			case <-interrupts:
				count++
				if count == 1 {
					cancelFirst()
				} else {
					cancelSecond()
				}
			case <-second.Done():
				return
			}
		}
	}()

	switch mode {
	case "run":
		return runMode(first, second, cfg)
	default:
		return cleanupMode(first, cfg)
	}
}

func parseArgs(args []string) (string, config, error) {
	cfg := defaultConfig()
	if len(args) == 0 || (args[0] != "run" && args[0] != "cleanup") {
		return "", cfg, errors.New("a mode of 'run' or 'cleanup' is required")
	}
	mode := args[0]
	flags := flag.NewFlagSet("agent-smoke "+mode, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.binary, "binary", "", "path to the zrok2 binary under test")
	flags.StringVar(&cfg.stateDir, "state-dir", "", "directory holding the run manifest and report")
	flags.StringVar(&cfg.apiEndpoint, "api-endpoint", cfg.apiEndpoint, "zrok v2 api endpoint (run only)")
	flags.StringVar(&cfg.namespace, "namespace", cfg.namespace, "namespace token for the reserved name (run only)")
	if err := flags.Parse(args[1:]); err != nil {
		return "", cfg, err
	}
	if flags.NArg() != 0 {
		return "", cfg, fmt.Errorf("unexpected arguments '%v'", flags.Args())
	}
	if cfg.stateDir == "" {
		return "", cfg, errors.New("--state-dir is required")
	}
	if mode == "run" && cfg.binary == "" {
		return "", cfg, errors.New("--binary is required")
	}
	if cfg.binary != "" {
		binary, err := filepath.Abs(cfg.binary)
		if err != nil {
			return "", cfg, err
		}
		cfg.binary = binary
	}
	cfg.accountToken = os.Getenv(accountTokenEnv)
	if value := os.Getenv("GITHUB_RUN_ID"); value != "" {
		cfg.runID = value
	}
	if value := os.Getenv("GITHUB_RUN_ATTEMPT"); value != "" {
		cfg.runAttempt = value
	}
	return mode, cfg, nil
}
