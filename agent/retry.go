package agent

import (
	"context"
	"math"
	"time"
)

func (a *Agent) startRetryLocked() {
	if a.retryStarted || a.stopping {
		return
	}
	a.retryStarted = true
	a.beginWorkLocked()
	go func() {
		defer a.endWork()
		interval := a.cfg.RetryCheckInterval
		if interval <= 0 {
			interval = 5 * time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-a.ctx.Done():
				return
			case now := <-ticker.C:
				a.retryDue(now)
			}
		}
	}()
}

func (a *Agent) retryDue(now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.stopping {
		return
	}
	for _, op := range a.operations {
		if op.failure == nil || op.failure.NextRetry.After(now) {
			continue
		}
		if op.phase == phaseFailed && (op.child != nil || op.resourceToken() != "") {
			a.claimCleanupLocked(op)
			continue
		}
		if op.cleanup != nil {
			select {
			case <-op.cleanup.done:
			default:
				continue
			}
		}
		if op.released || op.phase != phaseRetry || !op.desired {
			continue
		}
		if a.cfg.MaxRetries >= 0 && op.failure.Count >= a.cfg.MaxRetries {
			continue
		}
		a.claimStartLocked(op, context.Background(), "retry")
	}
}

// retryCalculator provides centralized exponential backoff calculation
// for retry operations throughout the agent system
type retryCalculator struct {
	initialDelay time.Duration
	maxDelay     time.Duration
}

// newRetryCalculator creates a new retry calculator with the specified delays
func newRetryCalculator(cfg *AgentConfig) *retryCalculator {
	return &retryCalculator{
		initialDelay: cfg.RetryInitialDelay,
		maxDelay:     cfg.RetryMaxDelay,
	}
}

// nextRetryTime returns the absolute time for the next retry
func (rc *retryCalculator) nextRetryTime(attemptCount int) time.Time {
	return time.Now().Add(rc.nextRetry(attemptCount))
}

// nextRetry calculates the next retry delay using exponential backoff
// attemptCount should be 1 for the first retry, 2 for the second, etc.
func (rc *retryCalculator) nextRetry(attemptCount int) time.Duration {
	if attemptCount <= 0 {
		return rc.initialDelay
	}

	// calculate exponential backoff: initialDelay * 2^(attemptCount-1)
	delay := float64(rc.initialDelay) * math.Pow(2, float64(attemptCount-1))

	// cap at maximum delay
	if delay > float64(rc.maxDelay) {
		return rc.maxDelay
	}

	return time.Duration(delay)
}

// shouldRetry checks if enough time has passed for a retry
func (rc *retryCalculator) shouldRetry(nextRetryTime time.Time) bool {
	return time.Now().After(nextRetryTime)
}
