//go:build agent_smoke

package main

import (
	"time"
)

const (
	accountTokenEnv  = "ZROK2_CI_ACCOUNT_TOKEN"
	defaultEndpoint  = "https://api-v2.zrok.io"
	defaultNamespace = "public"
)

// config carries the explicit inputs and every bound the harness enforces. tests shorten the bounds; the
// defaults are the work order's.
type config struct {
	binary       string
	stateDir     string
	apiEndpoint  string
	namespace    string
	accountToken string
	runID        string
	runAttempt   string

	scenarioBound time.Duration
	startupBound  time.Duration
	stopBound     time.Duration
	requestBound  time.Duration
	commandBound  time.Duration
	publicBound   time.Duration
	cleanupBound  time.Duration
	killBound     time.Duration
	pollInterval  time.Duration
}

func defaultConfig() config {
	return config{
		apiEndpoint:   defaultEndpoint,
		namespace:     defaultNamespace,
		runID:         "local",
		runAttempt:    "0",
		scenarioBound: 15 * time.Minute,
		startupBound:  7 * time.Minute,
		stopBound:     130 * time.Second,
		requestBound:  30 * time.Second,
		commandBound:  90 * time.Second,
		publicBound:   120 * time.Second,
		cleanupBound:  3 * time.Minute,
		killBound:     10 * time.Second,
		pollInterval:  2 * time.Second,
	}
}
