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
				dl.Errorf("failed to restart private access '%v': %v", access.Request.Token, err)
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
				delay := time.Duration(math.Min(
					float64(rm.a.cfg.RetryInitialDelay)*math.Pow(2, float64(access.Failure.Count-1)),
					float64(rm.a.cfg.RetryMaxDelay),
				))
				access.Failure.NextRetry = time.Now().Add(delay)
				registryModified = true
				newAccesses[failureId] = access

				dl.Infof("next retry for private access '%v' scheduled for '%v'", failureId, access.Failure.NextRetry)

			} else {
				access.Failure = nil
				registryModified = true
				dl.Infof("restarted private access '%v' -> '%v'", access.Request.Token, resp)
			}
		} else {
			newAccesses[failureId] = access
		}
	}
	rm.accesses = newAccesses

	newShares := make(map[string]*ShareRegistryEntry)
	for failureId, share := range rm.shares {
		if time.Now().After(share.Failure.NextRetry) {
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
				var delay = 30 * time.Second
				delay = time.Duration(math.Min(
					float64(rm.a.cfg.RetryInitialDelay)*math.Pow(2, float64(share.Failure.Count-1)),
					float64(rm.a.cfg.RetryMaxDelay),
				))
				share.Failure.NextRetry = time.Now().Add(delay)
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
	gen, err := nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 4)
	if err != nil {
		return "", err
	}
	return "err_" + gen(), nil
}
