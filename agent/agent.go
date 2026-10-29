package agent

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/agentUi"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/openziti/zrok/v2/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type agentDependencies struct {
	startChild     func(proctree.TailFunction, ...string) (*proctree.Child, error)
	deleteShare    func(context.Context, string) error
	deleteAccess   func(context.Context, string, string) error
	loadRegistry   func(string) (*Registry, error)
	saveRegistry   func(*Registry, string) error
	remoteListener func(string) (net.Listener, error)
}

type registryLoad struct {
	done chan struct{}
	err  error
}

type Agent struct {
	mu            sync.Mutex
	registryWrite sync.Mutex
	cfg           *AgentConfig
	root          env_core.Root
	ctx           context.Context
	cancel        context.CancelFunc
	options       lifecycleOptions
	deps          agentDependencies
	operations    map[string]*operation
	retryCalc     *retryCalculator
	workers       int
	idle          chan struct{}
	stopping      bool
	runStarted    bool
	retryStarted  bool
	replayStarted bool
	registryLoad  *registryLoad
	shutdownDone  chan struct{}
	shutdownErr   error
	agentSocket   string
	httpEndpoint  string
	listeners     []net.Listener
	grpcServers   []*grpc.Server
	httpServers   []*http.Server
}

func NewAgent(cfg *AgentConfig, root env_core.Root) (*Agent, error) {
	if !root.IsEnabled() {
		return nil, errors.New("unable to load environment; did you 'zrok2 enable'?")
	}
	if cfg == nil {
		cfg = DefaultConfig()
	}
	config := *cfg
	ctx, cancel := context.WithCancel(context.Background())
	idle := make(chan struct{})
	close(idle)
	a := &Agent{cfg: &config, root: root, ctx: ctx, cancel: cancel, options: defaultLifecycleOptions(), operations: map[string]*operation{}, retryCalc: newRetryCalculator(&config), idle: idle}
	a.deps = agentDependencies{
		startChild:  proctree.StartChild,
		deleteShare: func(_ context.Context, token string) error { return sdk.DeleteShare(root, &sdk.Share{Token: token}) },
		deleteAccess: func(_ context.Context, token, frontend string) error {
			return sdk.DeleteAccess(root, &sdk.Access{Token: frontend, ShareToken: token})
		},
		loadRegistry:   LoadRegistry,
		saveRegistry:   func(registry *Registry, path string) error { return registry.Save(path) },
		remoteListener: func(token string) (net.Listener, error) { return sdk.NewListener(token, root) },
	}
	return a, nil
}

func (a *Agent) Config() *AgentConfig { config := *a.cfg; return &config }

func (a *Agent) Run() error {
	err := a.run()
	return errors.Join(err, a.Shutdown())
}

func (a *Agent) run() error {
	a.mu.Lock()
	if a.stopping || a.runStarted {
		a.mu.Unlock()
		return errors.New("agent has already started or is stopping")
	}
	a.runStarted = true
	a.beginWorkLocked()
	a.mu.Unlock()
	if err := proctree.Init("zrok Agent"); err != nil {
		a.endWork()
		return err
	}
	socket, err := a.root.AgentSocket()
	if err != nil {
		a.endWork()
		return err
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		a.endWork()
		return err
	}
	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{agent: a})
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		_ = listener.Close()
		_ = os.Remove(socket)
		a.endWork()
		return nil
	}
	a.agentSocket = socket
	a.listeners = append(a.listeners, listener)
	a.grpcServers = append(a.grpcServers, srv)
	a.startRetryLocked()
	failures := make(chan error, 3)
	go func() {
		defer a.endWork()
		if err := srv.Serve(listener); err != nil && a.ctx.Err() == nil {
			failures <- err
		}
	}()
	if a.cfg.ConsoleEnabled {
		a.beginWorkLocked()
		go a.gateway(failures)
	}
	a.beginWorkLocked()
	go a.remoteAgent(failures)
	replay := make(chan error, 1)
	a.beginWorkLocked()
	go func() { defer a.endWork(); replay <- a.RestoreRegistryOnce() }()
	a.mu.Unlock()
	select {
	case err := <-replay:
		if err != nil {
			return err
		}
	case <-a.ctx.Done():
		return nil
	case err := <-failures:
		return err
	}
	select {
	case <-a.ctx.Done():
		return nil
	case err := <-failures:
		return err
	}
}

// Shutdown closes admission before arranging cleanup; repeated callers join the same result.
func (a *Agent) Shutdown() error {
	a.mu.Lock()
	if a.shutdownDone != nil {
		done := a.shutdownDone
		a.mu.Unlock()
		<-done
		return a.shutdownErr
	}
	a.shutdownDone = make(chan struct{})
	a.stopping = true
	a.cancel()
	a.ensureImportLocked()
	for _, op := range a.operations {
		op.stop = true
		if op.start != nil {
			op.start.cancel()
		}
		if op.child != nil || op.phase == phaseStarting || op.phase == phaseStopping || op.phase == phaseFailed && op.resourceToken() != "" {
			a.claimCleanupLocked(op)
		}
	}
	listeners := append([]net.Listener(nil), a.listeners...)
	servers := append([]*grpc.Server(nil), a.grpcServers...)
	httpServers := append([]*http.Server(nil), a.httpServers...)
	done := a.shutdownDone
	a.mu.Unlock()
	go a.finishShutdown(listeners, servers, httpServers)
	<-done
	return a.shutdownErr
}

func (a *Agent) finishShutdown(listeners []net.Listener, servers []*grpc.Server, httpServers []*http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), a.options.shutdown)
	defer cancel()
	stoppedServers := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for _, listener := range listeners {
			_ = listener.Close()
		}
		for _, srv := range servers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				stopped := make(chan struct{})
				go func() { srv.GracefulStop(); close(stopped) }()
				timer := time.NewTimer(a.options.server)
				defer timer.Stop()
				select {
				case <-stopped:
				case <-timer.C:
					srv.Stop()
				case <-ctx.Done():
					srv.Stop()
				}
			}()
		}
		for _, srv := range httpServers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bound, stop := context.WithTimeout(ctx, a.options.server)
				defer stop()
				if err := srv.Shutdown(bound); err != nil {
					_ = srv.Close()
				}
			}()
		}
		wg.Wait()
		close(stoppedServers)
	}()
	a.mu.Lock()
	idle := a.idle
	load := a.registryLoad
	a.mu.Unlock()
	var result error
	select {
	case <-idle:
	case <-ctx.Done():
		result = errors.Join(result, ctx.Err())
	}
	select {
	case <-stoppedServers:
	case <-ctx.Done():
		result = errors.Join(result, ctx.Err())
	}
	select {
	case <-load.done:
		result = errors.Join(result, load.err)
	case <-ctx.Done():
		result = errors.Join(result, ctx.Err())
	}
	a.mu.Lock()
	for _, op := range a.operations {
		if op.cleanup != nil {
			select {
			case <-op.cleanup.done:
				result = errors.Join(result, op.cleanup.err)
			default:
			}
		}
		if op.resourceToken() != "" && op.phase != phaseIdle {
			result = errors.Join(result, fmt.Errorf("incomplete cleanup for '%v'", op.resourceToken()))
		}
	}
	socket := a.agentSocket
	a.mu.Unlock()
	if result == nil || ctx.Err() == nil {
		saved := make(chan error, 1)
		go func() { saved <- a.SaveRegistry() }()
		select {
		case err := <-saved:
			result = errors.Join(result, err)
		case <-ctx.Done():
			result = errors.Join(result, ctx.Err())
		}
	}
	if socket != "" {
		if err := os.Remove(socket); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, err)
		}
	}
	a.mu.Lock()
	a.shutdownErr = result
	close(a.shutdownDone)
	a.mu.Unlock()
}

func (a *Agent) remoteAgent(failures chan<- error) {
	defer a.endWork()
	path, err := a.root.AgentEnrollment()
	if err == nil {
		var enrollment *Enrollment
		enrollment, err = LoadEnrollment(path)
		if err == nil {
			var listener net.Listener
			listener, err = a.deps.remoteListener(enrollment.Token)
			if err == nil {
				srv := grpc.NewServer()
				agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{agent: a})
				a.mu.Lock()
				if a.stopping {
					a.mu.Unlock()
					_ = listener.Close()
					return
				}
				a.listeners = append(a.listeners, listener)
				a.grpcServers = append(a.grpcServers, srv)
				a.mu.Unlock()
				err = srv.Serve(listener)
			}
		}
	}
	if err != nil && a.ctx.Err() == nil {
		if a.cfg.RequireRemoting {
			failures <- fmt.Errorf("required agent remoting: %w", err)
		} else {
			dl.Warnf("agent remoting unavailable: %v", err)
		}
	}
}

func (a *Agent) gateway(failures chan<- error) {
	defer a.endWork()
	mux := runtime.NewServeMux()
	a.mu.Lock()
	socket := a.agentSocket
	a.mu.Unlock()
	if err := agentGrpc.RegisterAgentHandlerFromEndpoint(a.ctx, mux, "unix:"+socket, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}); err != nil {
		if a.ctx.Err() == nil {
			failures <- err
		}
		return
	}
	listener, err := util.AutoListener("tcp", a.cfg.ConsoleAddress, a.cfg.ConsoleStartPort, a.cfg.ConsoleEndPort)
	if err != nil {
		if a.ctx.Err() == nil {
			failures <- err
		}
		return
	}
	srv := &http.Server{Handler: agentUi.Middleware(mux)}
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		_ = listener.Close()
		return
	}
	a.httpEndpoint = listener.Addr().String()
	a.listeners = append(a.listeners, listener)
	a.httpServers = append(a.httpServers, srv)
	a.mu.Unlock()
	if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) && a.ctx.Err() == nil {
		failures <- err
	}
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	agent *Agent
}
