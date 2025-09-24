package agent

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/agentUi"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Agent struct {
	cfg             *AgentConfig
	httpEndpoint    string
	root            env_core.Root
	agentSocket     string
	shares          map[string]*share
	addShare        chan *share
	rmShare         chan *share
	accesses        map[string]*access
	addAccess       chan *access
	rmAccess        chan *access
	persistRegistry bool
	retryTicker     *time.Ticker
	stopRetries     chan bool
	failedAccesses  map[string]*AccessRegistryEntry
	failedShares    map[string]*ShareRegistryEntry
	nextFailureID   int64
}

// generateSessionFailureID creates a unique ID for this agent session
func (a *Agent) generateSessionFailureID() string {
	id := atomic.AddInt64(&a.nextFailureID, 1)
	return fmt.Sprintf("failure_%d", id)
}

func NewAgent(cfg *AgentConfig, root env_core.Root) (*Agent, error) {
	if !root.IsEnabled() {
		return nil, errors.Errorf("unable to load environment; did you 'zrok enable'?")
	}
	return &Agent{
		cfg:            cfg,
		root:           root,
		shares:         make(map[string]*share),
		addShare:       make(chan *share),
		rmShare:        make(chan *share),
		accesses:       make(map[string]*access),
		addAccess:      make(chan *access),
		rmAccess:       make(chan *access),
		stopRetries:    make(chan bool),
		failedAccesses: make(map[string]*AccessRegistryEntry),
		failedShares:   make(map[string]*ShareRegistryEntry),
		nextFailureID:  0,
	}, nil
}

func (a *Agent) Run() error {
	logrus.Infof("started")

	if err := proctree.Init("zrok Agent"); err != nil {
		return err
	}

	agentSocket, err := a.root.AgentSocket()
	if err != nil {
		return err
	}
	l, err := net.Listen("unix", agentSocket)
	if err != nil {
		return err
	}
	a.agentSocket = agentSocket

	go a.manager()
	if a.cfg.ConsoleEnabled {
		go a.gateway()
	}
	go a.remoteAgent()

	a.persistRegistry = false
	if err := a.ReloadRegistry(); err != nil {
		logrus.Errorf("error reloading registry '%v'", err)
	}
	a.persistRegistry = true

	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{agent: a})
	if err := srv.Serve(l); err != nil {
		return err
	}

	return nil
}

func (a *Agent) Shutdown() {
	logrus.Infof("stopping")

	a.persistRegistry = false

	// stop retry manager
	if a.stopRetries != nil {
		select {
		case a.stopRetries <- true:
			logrus.Debug("signaled retry manager to stop")
		default:
			logrus.Debug("retry manager already stopped or stopping")
		}
	}

	if err := os.Remove(a.agentSocket); err != nil {
		logrus.Warnf("unable to remove agent socket: %v", err)
	}
	for _, shr := range a.shares {
		logrus.Debugf("stopping share '%v'", shr.token)
		a.rmShare <- shr
	}
	for _, acc := range a.accesses {
		logrus.Debugf("stopping access '%v'", acc.token)
		a.rmAccess <- acc
	}
}

func (a *Agent) Config() *AgentConfig {
	return a.cfg
}

func (a *Agent) ReloadRegistry() error {
	registryPath, err := a.root.AgentRegistry()
	if err != nil {
		return err
	}
	registry, err := LoadRegistry(registryPath)
	if err != nil {
		return err
	}

	logrus.Infof("loaded %d accesses", len(registry.PrivateAccesses))
	registryModified := false
	for _, entry := range registry.PrivateAccesses {
		if resp, err := a.AccessPrivate(entry.Request); err == nil {
			logrus.Infof("restarted private access '%v' -> '%v'", entry.Request.Token, resp)
			// reset failure state on success
			if entry.FailureCount > 0 {
				entry.FailureCount = 0
				entry.LastFailure = nil
				entry.LastError = ""
				entry.NextRetry = nil
				registryModified = true
			}
		} else {
			logrus.Warnf("failed to restart private access '%v': %v (will retry)", entry.Request.Token, err)
			entry.FailureCount++
			now := time.Now()
			entry.LastFailure = &now
			entry.LastError = err.Error()

			// calculate next retry with exponential backoff
			delay := time.Duration(math.Min(
				float64(a.cfg.RetryInitialDelay)*math.Pow(2, float64(entry.FailureCount-1)),
				float64(a.cfg.RetryMaxDelay),
			))
			nextRetry := now.Add(delay)
			entry.NextRetry = &nextRetry
			registryModified = true

			logrus.Infof("next retry for private access '%v' scheduled for %v",
				entry.Request.Token, nextRetry.Format(time.RFC3339))
		}
	}

	logrus.Infof("loaded %d public shares", len(registry.PublicShares))
	for _, entry := range registry.PublicShares {
		if token, frontends, err := a.SharePublic(entry.Request); err == nil {
			logrus.Infof("restarted public share '%v' -> token='%v', frontends='%v'", entry.Request.Target, token, frontends)
			// reset failure state on success
			if entry.FailureCount > 0 {
				entry.FailureCount = 0
				entry.LastFailure = nil
				entry.LastError = ""
				entry.NextRetry = nil
				registryModified = true
			}
		} else {
			logrus.Warnf("failed to restart public share '%v': %v (will retry)", entry.Request.Target, err)
			entry.FailureCount++
			now := time.Now()
			entry.LastFailure = &now
			entry.LastError = err.Error()

			// calculate next retry with exponential backoff
			delay := time.Duration(math.Min(
				float64(a.cfg.RetryInitialDelay)*math.Pow(2, float64(entry.FailureCount-1)),
				float64(a.cfg.RetryMaxDelay),
			))
			nextRetry := now.Add(delay)
			entry.NextRetry = &nextRetry
			registryModified = true

			logrus.Infof("next retry for public share '%v' scheduled for %v", entry.Request.Target, nextRetry.Format(time.RFC3339))
		}
	}

	// save updated registry with retry state
	if registryModified {
		if err := registry.Save(registryPath); err != nil {
			logrus.Errorf("error saving updated registry: %v", err)
		}
	}

	// populate failed item maps with session IDs for status tracking
	a.failedAccesses = make(map[string]*AccessRegistryEntry)
	a.failedShares = make(map[string]*ShareRegistryEntry)

	for _, entry := range registry.PrivateAccesses {
		if entry.NextRetry != nil { // has retry state = failed
			failureID := a.generateSessionFailureID()
			a.failedAccesses[failureID] = entry
		}
	}

	for _, entry := range registry.PublicShares {
		if entry.NextRetry != nil { // has retry state = failed
			failureID := a.generateSessionFailureID()
			a.failedShares[failureID] = entry
		}
	}

	// start retry manager if we have any failed shares or accesses
	hasFailedItems := len(a.failedAccesses) > 0 || len(a.failedShares) > 0
	if hasFailedItems {
		go a.retryManager()
	}

	logrus.Infof("reload complete")
	return nil
}

func (a *Agent) SaveRegistry() error {
	r := &Registry{}
	// save private accesses with retry state
	for _, acc := range a.accesses {
		if acc.request != nil {
			// create new registry entry for this access
			entry := &AccessRegistryEntry{
				Request: acc.request,
				// retry state will be set by retry logic if needed
			}
			r.PrivateAccesses = append(r.PrivateAccesses, entry)
		}
	}
	// save public shares with registered names (namespace:name format)
	for _, shr := range a.shares {
		if req, ok := shr.request.(*SharePublicRequest); ok {
			// only save shares with at least one registered name (not just namespace)
			hasRegisteredName := false
			for _, ns := range req.NameSelections {
				if ns.Name != "" {
					hasRegisteredName = true
					break
				}
			}
			if hasRegisteredName {
				// create new registry entry for this share
				entry := &ShareRegistryEntry{
					Request: req,
					// retry state will be set by retry logic if needed
				}
				r.PublicShares = append(r.PublicShares, entry)
			}
		}
	}
	registryPath, err := a.root.AgentRegistry()
	if err != nil {
		return err
	}
	if err := r.Save(registryPath); err != nil {
		return err
	}
	return nil
}

func (a *Agent) remoteAgent() {
	enrollmentPath, err := a.root.AgentEnrollment()
	if err != nil {
		logrus.Errorf("unable to get agent enrollment path: %v", err)
		return
	}

	enrollment, err := LoadEnrollment(enrollmentPath)
	if err != nil {
		logrus.Warnf("unable to load agent enrollment: %v", err)
		return
	}

	logrus.Infof("listening for remote commands at '%v'", enrollment.Token)

	l, err := sdk.NewListener(enrollment.Token, a.root)
	if err != nil {
		logrus.Errorf("error listening for remote agent: %v", err)
		return
	}

	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{agent: a})
	if err := srv.Serve(l); err != nil {
		logrus.Errorf("error serving: %v", err)
		return
	}
}

func (a *Agent) gateway() {
	logrus.Info("started")
	defer logrus.Warn("exited")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := "unix:" + a.agentSocket
	logrus.Debugf("endpoint: '%v'", endpoint)
	if err := agentGrpc.RegisterAgentHandlerFromEndpoint(ctx, mux, "unix:"+a.agentSocket, opts); err != nil {
		logrus.Fatalf("unable to register gateway: %v", err)
	}

	listener, err := util.AutoListener("tcp", a.cfg.ConsoleAddress, a.cfg.ConsoleStartPort, a.cfg.ConsoleEndPort)
	if err != nil {
		logrus.Fatalf("unable to create a listener: %v", err)
	}
	a.httpEndpoint = listener.Addr().String()

	if err := http.Serve(listener, agentUi.Middleware(mux)); err != nil {
		logrus.Error(err)
	}
}

func (a *Agent) manager() {
	logrus.Info("started")
	defer logrus.Warn("exited")

	for {
		select {
		case inShare := <-a.addShare:
			logrus.Infof("adding new share '%v'", inShare.token)
			a.shares[inShare.token] = inShare

			if a.persistRegistry {
				if err := a.SaveRegistry(); err != nil {
					logrus.Errorf("unable to persist registry: %v", err)
				}
			}

		case outShare := <-a.rmShare:
			if shr, found := a.shares[outShare.token]; found {
				logrus.Infof("removing share '%v'", shr.token)
				if err := proctree.StopChild(shr.process); err != nil {
					logrus.Errorf("error stopping share '%v': %v", shr.token, err)
				}
				if err := proctree.WaitChild(shr.process); err != nil {
					logrus.Errorf("error joining share '%v': %v", shr.token, err)
				}
				if err := a.deleteShare(shr.token); err != nil {
					logrus.Errorf("error deleting share '%v': %v", shr.token, err)
				}
				delete(a.shares, shr.token)

				if a.persistRegistry {
					if err := a.SaveRegistry(); err != nil {
						logrus.Errorf("unable to persist registry: %v", err)
					}
				}

			} else {
				logrus.Debug("skipping unidentified (orphaned) share removal")
			}

		case inAccess := <-a.addAccess:
			logrus.Infof("adding new access '%v'", inAccess.frontendToken)
			a.accesses[inAccess.frontendToken] = inAccess

			if a.persistRegistry {
				if err := a.SaveRegistry(); err != nil {
					logrus.Errorf("unable to persist registry: %v", err)
				}
			}

		case outAccess := <-a.rmAccess:
			if acc, found := a.accesses[outAccess.frontendToken]; found {
				logrus.Infof("removing access '%v'", acc.frontendToken)
				if err := proctree.StopChild(acc.process); err != nil {
					logrus.Errorf("error stopping access '%v': %v", acc.frontendToken, err)
				}
				if err := proctree.WaitChild(acc.process); err != nil {
					logrus.Errorf("error joining access '%v': %v", acc.frontendToken, err)
				}
				if err := a.deleteAccess(acc.token, acc.frontendToken); err != nil {
					logrus.Errorf("error deleting access '%v': %v", acc.frontendToken, err)
				}
				delete(a.accesses, acc.frontendToken)

				if a.persistRegistry {
					if err := a.SaveRegistry(); err != nil {
						logrus.Errorf("unable to persist registry: %v", err)
					}
				}

			} else {
				logrus.Debug("skipping unidentified (orphaned) access removal")
			}
		}
	}
}

func (a *Agent) deleteShare(token string) error {
	logrus.Debugf("deleting share '%v'", token)
	if err := sdk.DeleteShare(a.root, &sdk.Share{Token: token}); err != nil {
		return err
	}
	return nil
}

func (a *Agent) deleteAccess(token, frontendToken string) error {
	logrus.Debugf("deleting access '%v'", frontendToken)
	if err := sdk.DeleteAccess(a.root, &sdk.Access{Token: frontendToken, ShareToken: token}); err != nil {
		return err
	}
	return nil
}

func (a *Agent) retryManager() {
	logrus.Info("retry manager started")
	defer logrus.Info("retry manager stopped")

	a.retryTicker = time.NewTicker(a.cfg.RetryCheckInterval)
	defer a.retryTicker.Stop()

	for {
		select {
		case <-a.retryTicker.C:
			a.processRetries()
		case <-a.stopRetries:
			return
		}
	}
}

func (a *Agent) processRetries() {
	registryPath, err := a.root.AgentRegistry()
	if err != nil {
		logrus.Errorf("unable to get agent registry path: %v", err)
		return
	}

	registry, err := LoadRegistry(registryPath)
	if err != nil {
		logrus.Errorf("unable to load registry for retry processing: %v", err)
		return
	}

	now := time.Now()
	registryModified := false
	activeRetries := false

	// process private accesses that are ready for retry - rebuild failed map
	newFailedAccesses := make(map[string]*AccessRegistryEntry)
	var activeAccessEntries []*AccessRegistryEntry
	for _, entry := range registry.PrivateAccesses {
		if entry.NextRetry != nil && now.After(*entry.NextRetry) {
			if a.cfg.MaxRetries > -1 && entry.FailureCount >= a.cfg.MaxRetries {
				logrus.Warnf("abandoning private access '%v' after %d failed attempts, last error: %v",
					entry.Request.Token, entry.FailureCount, entry.LastError)
				registryModified = true
				continue // abandon this access
			}

			if resp, err := a.AccessPrivate(entry.Request); err == nil {
				logrus.Infof("retry succeeded for private access '%v' -> '%v' (recovered from: %v)",
					entry.Request.Token, resp, entry.LastError)
				// clear failure state
				entry.FailureCount = 0
				entry.LastFailure = nil
				entry.LastError = ""
				entry.NextRetry = nil
				registryModified = true
				// SUCCESS: don't add to newFailedAccesses (removed from failed map)
			} else {
				logrus.Warnf("retry %d failed for private access '%v': %v",
					entry.FailureCount+1, entry.Request.Token, err)
				entry.FailureCount++
				entry.LastFailure = &now
				entry.LastError = err.Error()

				// exponential backoff
				delay := time.Duration(math.Min(
					float64(a.cfg.RetryInitialDelay)*math.Pow(2, float64(entry.FailureCount-1)),
					float64(a.cfg.RetryMaxDelay),
				))
				nextRetry := now.Add(delay)
				entry.NextRetry = &nextRetry
				registryModified = true
				activeRetries = true

				logrus.Debugf("next retry for private access '%v' scheduled for %v",
					entry.Request.Token, nextRetry.Format(time.RFC3339))

				// find existing failure ID or create new one
				var failureID string
				for fid, existingEntry := range a.failedAccesses {
					if existingEntry.Request.Token == entry.Request.Token {
						failureID = fid
						break
					}
				}
				if failureID == "" {
					failureID = a.generateSessionFailureID()
				}
				newFailedAccesses[failureID] = entry
			}
		} else if entry.NextRetry != nil {
			activeRetries = true
			// find existing failure ID or create new one
			var failureID string
			for fid, existingEntry := range a.failedAccesses {
				if existingEntry.Request.Token == entry.Request.Token {
					failureID = fid
					break
				}
			}
			if failureID == "" {
				failureID = a.generateSessionFailureID()
			}
			newFailedAccesses[failureID] = entry
		}
		activeAccessEntries = append(activeAccessEntries, entry)
	}
	a.failedAccesses = newFailedAccesses

	// process public shares that are ready for retry - rebuild failed map
	newFailedShares := make(map[string]*ShareRegistryEntry)
	var activeShareEntries []*ShareRegistryEntry
	for _, entry := range registry.PublicShares {
		if entry.NextRetry != nil && now.After(*entry.NextRetry) {
			if a.cfg.MaxRetries > -1 && entry.FailureCount >= a.cfg.MaxRetries {
				logrus.Warnf("abandoning public share '%v' after %d failed attempts, last error: %v",
					entry.Request.Target, entry.FailureCount, entry.LastError)
				// skip this entry (don't add to activeShareEntries)
				registryModified = true
				continue
			}

			if token, _, err := a.SharePublic(entry.Request); err == nil {
				logrus.Infof("retry succeeded for public share '%v' -> token='%v' (recovered from: %v)",
					entry.Request.Target, token, entry.LastError)
				// clear all failure state
				entry.FailureCount = 0
				entry.LastFailure = nil
				entry.LastError = ""
				entry.NextRetry = nil
				registryModified = true
				// SUCCESS: don't add to newFailedShares (removed from failed map)
			} else {
				logrus.Warnf("retry %d failed for public share '%v': %v",
					entry.FailureCount+1, entry.Request.Target, err)
				entry.FailureCount++
				entry.LastFailure = &now
				entry.LastError = err.Error()

				// calculate next retry with exponential backoff
				delay := time.Duration(math.Min(
					float64(a.cfg.RetryInitialDelay)*math.Pow(2, float64(entry.FailureCount-1)),
					float64(a.cfg.RetryMaxDelay),
				))
				nextRetry := now.Add(delay)
				entry.NextRetry = &nextRetry
				registryModified = true
				activeRetries = true

				logrus.Debugf("next retry for public share '%v' scheduled for %v",
					entry.Request.Target, nextRetry.Format(time.RFC3339))

				// find existing failure ID or create new one
				var failureID string
				for fid, existingEntry := range a.failedShares {
					if existingEntry.Request.Target == entry.Request.Target &&
						len(existingEntry.Request.NameSelections) == len(entry.Request.NameSelections) {
						// match by target and name selections to identify same request
						match := true
						for i, ns := range existingEntry.Request.NameSelections {
							if i < len(entry.Request.NameSelections) &&
								(ns.NamespaceToken != entry.Request.NameSelections[i].NamespaceToken ||
									ns.Name != entry.Request.NameSelections[i].Name) {
								match = false
								break
							}
						}
						if match {
							failureID = fid
							break
						}
					}
				}
				if failureID == "" {
					failureID = a.generateSessionFailureID()
				}
				newFailedShares[failureID] = entry
			}
		} else if entry.NextRetry != nil {
			// still has pending retries
			activeRetries = true
			// find existing failure ID or create new one
			var failureID string
			for fid, existingEntry := range a.failedShares {
				if existingEntry.Request.Target == entry.Request.Target &&
					len(existingEntry.Request.NameSelections) == len(entry.Request.NameSelections) {
					// match by target and name selections to identify same request
					match := true
					for i, ns := range existingEntry.Request.NameSelections {
						if i < len(entry.Request.NameSelections) &&
							(ns.NamespaceToken != entry.Request.NameSelections[i].NamespaceToken ||
								ns.Name != entry.Request.NameSelections[i].Name) {
							match = false
							break
						}
					}
					if match {
						failureID = fid
						break
					}
				}
			}
			if failureID == "" {
				failureID = a.generateSessionFailureID()
			}
			newFailedShares[failureID] = entry
		}
		activeShareEntries = append(activeShareEntries, entry)
	}
	a.failedShares = newFailedShares

	// update registry with active entries (abandoned items removed)
	registry.PrivateAccesses = activeAccessEntries
	registry.PublicShares = activeShareEntries

	// save registry if modified
	if registryModified {
		if err := registry.Save(registryPath); err != nil {
			logrus.Errorf("error saving registry after retry processing: %v", err)
		}
	}

	// if no more active retries, stop the retry manager
	if !activeRetries {
		logrus.Debug("no more shares need retries, stopping retry manager")
		go func() {
			select {
			case a.stopRetries <- true:
			default:
			}
		}()
	}
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	agent *Agent
}
