//go:build agent_smoke

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/openziti/zrok/v2/agent/agentClient"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/environment/env_core"
	"google.golang.org/grpc"
)

func randomHex(bytes int) string {
	buf := make([]byte, bytes)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// pause waits one poll interval, failing early when the deadline passes or the agent exits.
func pause(ctx context.Context, agent *child, interval time.Duration, last error) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		if last != nil {
			return fmt.Errorf("%w; last observation: %v", ctx.Err(), last)
		}
		return ctx.Err()
	case <-agent.done:
		return fmt.Errorf("agent exited with status %d; last observation: %v", agent.exitCode(), last)
	case <-timer.C:
		return nil
	}
}

// awaitVersion polls the agent socket until the agent answers, then requires the expected build identity. each
// attempt uses a fresh connection, so reconnect backoff never stretches the polling interval.
func awaitVersion(ctx context.Context, cfg config, root env_core.Root, agent *child, expected string) (agentGrpc.AgentClient, *grpc.ClientConn, error) {
	var last error
	for {
		client, conn, err := agentClient.NewClient(root)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create agent client: %w", err)
		}
		callCtx, cancel := context.WithTimeout(ctx, cfg.requestBound)
		resp, err := client.Version(callCtx, &agentGrpc.VersionRequest{})
		cancel()
		if err == nil {
			if resp.GetV() != expected {
				_ = conn.Close()
				return nil, nil, fmt.Errorf("agent reported build '%v'; expected '%v'", resp.GetV(), expected)
			}
			return client, conn, nil
		}
		_ = conn.Close()
		last = err
		if err := pause(ctx, agent, cfg.pollInterval, last); err != nil {
			return nil, nil, fmt.Errorf("agent did not answer its version request: %w", err)
		}
	}
}

// awaitShare polls structured agent status until exactly one public share is active. every token seen is
// recorded first. more than one share, any access, or a failed share ends the wait at once.
func awaitShare(ctx context.Context, cfg config, client agentGrpc.AgentClient, agent *child, record func(string) error) (*agentGrpc.ShareDetail, error) {
	var last error
	for {
		callCtx, cancel := context.WithTimeout(ctx, cfg.requestBound)
		resp, err := client.Status(callCtx, &agentGrpc.StatusRequest{})
		cancel()
		if err == nil {
			for _, shr := range resp.GetShares() {
				if shr.GetToken() != "" {
					if err := record(shr.GetToken()); err != nil {
						return nil, err
					}
				}
			}
			detail, settled, err := activeShare(resp)
			if err != nil {
				return nil, err
			}
			if settled {
				return detail, nil
			}
			last = errors.New(describeStatus(resp))
		} else {
			last = err
		}
		if err := pause(ctx, agent, cfg.pollInterval, last); err != nil {
			return nil, fmt.Errorf("no active share: %w", err)
		}
	}
}

func activeShare(resp *agentGrpc.StatusResponse) (*agentGrpc.ShareDetail, bool, error) {
	if len(resp.GetAccesses()) != 0 {
		return nil, false, fmt.Errorf("agent status lists %d accesses; expected none", len(resp.GetAccesses()))
	}
	shares := resp.GetShares()
	switch {
	case len(shares) == 0:
		return nil, false, nil
	case len(shares) > 1:
		return nil, false, fmt.Errorf("agent status lists %d shares; expected exactly one (%v)", len(shares), describeStatus(resp))
	}
	shr := shares[0]
	if shr.GetShareMode() != "public" {
		return nil, false, fmt.Errorf("agent share has mode '%v'; expected 'public'", shr.GetShareMode())
	}
	if shr.GetStatus() == "failed" {
		return nil, false, fmt.Errorf("agent share failed: %v", shr.GetFailure().GetLastError())
	}
	if shr.GetStatus() != "active" || shr.GetToken() == "" || len(shr.GetFrontendEndpoint()) == 0 {
		return nil, false, nil
	}
	return shr, true, nil
}

func describeStatus(resp *agentGrpc.StatusResponse) string {
	var parts []string
	for _, shr := range resp.GetShares() {
		part := fmt.Sprintf("share '%v' status '%v'", shr.GetToken(), shr.GetStatus())
		if lastError := shr.GetFailure().GetLastError(); lastError != "" {
			part += fmt.Sprintf(" error '%v'", lastError)
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return "no shares"
	}
	return strings.Join(parts, "; ")
}

// loopback serves one random body on one random path; every other path is not found.
type loopback struct {
	server *http.Server
	addr   string
	path   string
	body   string
}

func startLoopback() (*loopback, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	lb := &loopback{addr: listener.Addr().String(), path: "/" + randomHex(16), body: randomHex(32)}
	mux := http.NewServeMux()
	mux.HandleFunc(lb.path, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, lb.body)
	})
	lb.server = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = lb.server.Serve(listener) }()
	return lb, nil
}

func (lb *loopback) target() string {
	return "http://" + lb.addr
}

func (lb *loopback) close() {
	_ = lb.server.Close()
}

// awaitBody fetches the random path through a public endpoint until it returns the exact random body.
// named endpoints can be bare hostnames; assume https when no scheme is supplied.
func awaitBody(ctx context.Context, cfg config, endpoint string, lb *loopback) (string, error) {
	if !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}
	url := strings.TrimSuffix(endpoint, "/") + lb.path
	client := &http.Client{Timeout: cfg.requestBound}
	var last error
	for {
		last = fetchBody(ctx, client, url, lb.body)
		if last == nil {
			return url, nil
		}
		timer := time.NewTimer(cfg.pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", fmt.Errorf("'%v' did not return the expected body: %w; last observation: %v", url, ctx.Err(), last)
		case <-timer.C:
		}
	}
}

func fetchBody(ctx context.Context, client *http.Client, url, expected string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("skip_zrok_interstitial", "true")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(len(expected))+1024))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK || string(body) != expected {
		return fmt.Errorf("status %d with %d bytes not matching the expected body", resp.StatusCode, len(body))
	}
	return nil
}
