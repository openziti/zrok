package agent

import (
	"math"
	"time"

	"github.com/jaevor/go-nanoid"
	"github.com/michaelquigley/df/dl"
)

type retryManager struct {
	a         *Agent
	close     chan struct{}
	addAccess chan *AccessRegistryEntry
	rmAccess  chan string
	accesses  map[string]*AccessRegistryEntry
	addShare  chan *ShareRegistryEntry
	rmShare   chan string
	shares    map[string]*ShareRegistryEntry
	retryCalc *retryCalculator
}

func newRetryManager(a *Agent) *retryManager {
	return &retryManager{
		a:         a,
		close:     make(chan struct{}),
		addAccess: make(chan *AccessRegistryEntry),
		rmAccess:  make(chan string),
		accesses:  make(map[string]*AccessRegistryEntry),
		addShare:  make(chan *ShareRegistryEntry),
		rmShare:   make(chan string),
		shares:    make(map[string]*ShareRegistryEntry),
		retryCalc: newRetryCalculator(a.cfg),
	}
}

func (rm *retryManager) stop() {
	close(rm.close)
}

func (rm *retryManager) addFailedAccess(failed *AccessRegistryEntry) {
	rm.addAccess <- failed
}

func (rm *retryManager) hasFailedAccess(failureId string) bool {
	_, found := rm.accesses[failureId]
	return found
}

func (rm *retryManager) rmFailedAccess(failureId string) {
	rm.rmAccess <- failureId
}

func (rm *retryManager) addFailedShare(failed *ShareRegistryEntry) {
	rm.addShare <- failed
}

func (rm *retryManager) hasFailedShare(failureId string) bool {
	_, found := rm.shares[failureId]
	return found
}

func (rm *retryManager) rmFailedShare(failureId string) {
	rm.rmShare <- failureId
}

func (rm *retryManager) failures() ([]*AccessRegistryEntry, []*ShareRegistryEntry) {
	var accesses []*AccessRegistryEntry
	for _, access := range rm.accesses {
		accesses = append(accesses, access)
	}
	var shares []*ShareRegistryEntry
	for _, share := range rm.shares {
		shares = append(shares, share)
	}
	return accesses, shares
}

func (rm *retryManager) run() {
	dl.Info("started")
	defer dl.Info("exited")

	for {
		select {
		case <-rm.close:
			return

		case failedAccess := <-rm.addAccess:
			if failureId, err := rm.generateId(); err == nil {
				rm.accesses[failureId] = failedAccess
				dl.Infof("added access failure with id '%v'", failureId)
			} else {
				dl.Errorf("error adding access failure: %v", err)
			}

			if err := rm.a.SaveRegistry(); err != nil {
				dl.Errorf("error saving registry: %v", err)
			}

		case failureId := <-rm.rmAccess:
			delete(rm.accesses, failureId)

			if err := rm.a.SaveRegistry(); err != nil {
				dl.Errorf("error saving registry: %v", err)
			}

		case failedShare := <-rm.addShare:
			if shareId, err := rm.generateId(); err == nil {
				rm.shares[shareId] = failedShare
				dl.Infof("added share with id '%v'", shareId)
			} else {
				dl.Errorf("error adding access failure: %v", err)
			}

			if err := rm.a.SaveRegistry(); err != nil {
				dl.Errorf("error saving registry: %v", err)
			}

		case failureId := <-rm.rmShare:
			delete(rm.shares, failureId)

			if err := rm.a.SaveRegistry(); err != nil {
				dl.Errorf("error saving registry: %v", err)
			}

		case <-time.After(5 * time.Second):
			rm.retry()
		}
	}
}

func (rm *retryManager) retry() {
	dl.Debug("retrying")
	defer dl.Debug("exiting")

	registryModified := false
	newAccesses := make(map[string]*AccessRegistryEntry)
	for failureId, access := range rm.accesses {
		if time.Now().After(access.Failure.NextRetry) {
			if resp, err := rm.a.AccessPrivate(access.Request); err != nil {
				dl.Errorf("failed to restart private access '%v': %v", access.Request.ShareToken, err)
				if access.Failure != nil {
					access.Failure.Count++
					access.Failure.LastError = err.Error()
				} else {
					access.Failure = &FailureEntry{
						Count:     1,
						LastError: err.Error(),
					}
				}

				// calculate next retry with exponential backoff
				access.Failure.NextRetry = rm.retryCalc.nextRetryTime(access.Failure.Count)
				registryModified = true
				newAccesses[failureId] = access

				dl.Infof("next retry for private access '%v' scheduled for '%v'", failureId, access.Failure.NextRetry)

			} else {
				access.Failure = nil
				registryModified = true
				dl.Infof("restarted private access '%v' -> '%v'", access.Request.ShareToken, resp)
			}
		} else {
			newAccesses[failureId] = access
		}
	}
	rm.accesses = newAccesses

	newShares := make(map[string]*ShareRegistryEntry)
	for failureId, share := range rm.shares {
		if rm.retryCalc.shouldRetry(share.Failure.NextRetry) {
			if shrToken, fes, err := rm.a.SharePublic(share.Request); err != nil {
				dl.Errorf("failed to restart public share '%v': %v", failureId, err)
				if share.Failure != nil {
					share.Failure.Count++
					share.Failure.LastError = err.Error()
				} else {
					share.Failure = &FailureEntry{
						Count:     1,
						LastError: err.Error(),
					}
				}

				// calculate next retry with exponential backoff
				share.Failure.NextRetry = rm.retryCalc.nextRetryTime(share.Failure.Count)
				registryModified = true
				newShares[failureId] = share

				dl.Infof("next retry for public share '%v' scheduled for '%v'", failureId, share.Failure.NextRetry)

			} else {
				share.Failure = nil
				registryModified = true
				dl.Infof("restarted public share '%v' -> '%v'", shrToken, fes)
			}
		} else {
			newShares[failureId] = share
		}
	}
	rm.shares = newShares

	if registryModified {
		rm.a.SaveRegistry()
	}
}

func (rm *retryManager) generateId() (string, error) {
	gen, err := nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 8)
	if err != nil {
		return "", err
	}
	return "err_" + gen(), nil
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
