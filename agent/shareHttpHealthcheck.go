package agent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

func (a *Agent) ShareHttpHealthcheck(shareToken, endpoint, httpVerb string, expectedHttpResponse, timeoutMs int) error {
	a.mu.Lock()
	var snapshot *share
	for _, op := range a.operations {
		if op.phase == phaseActive && op.shr != nil && op.shr.token == shareToken {
			snapshot = &share{target: op.shr.target, backendMode: op.shr.backendMode}
			break
		}
	}
	a.mu.Unlock()
	if snapshot == nil {
		return fmt.Errorf("share '%v' not found in agent", shareToken)
	}
	if snapshot.backendMode != sdk.ProxyBackendMode {
		return fmt.Errorf("cannot perform http healthcheck on '%v' share '%v'", snapshot.backendMode, shareToken)
	}
	return a.doHttpHealthcheck(snapshot, endpoint, httpVerb, expectedHttpResponse, timeoutMs)
}

func (a *Agent) doHttpHealthcheck(shr *share, endpoint, httpVerb string, expectedHttpResponse, timeoutMs int) error {
	url := fmt.Sprintf("%v%v", shr.target, endpoint)
	req, err := http.NewRequest(httpVerb, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	timeout := 5 * time.Second
	if timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedHttpResponse {
		return fmt.Errorf("unexpected status code; got '%v', want '%v'", resp.StatusCode, expectedHttpResponse)
	}
	return nil
}

func (i *agentGrpcImpl) ShareHttpHealthcheck(_ context.Context, req *agentGrpc.ShareHttpHealthcheckRequest) (*agentGrpc.ShareHttpHealthcheckResponse, error) {
	if err := i.agent.ShareHttpHealthcheck(req.Token, req.Endpoint, req.HttpVerb, int(req.ExpectedHttpResponse), int(req.TimeoutMs)); err != nil {
		return &agentGrpc.ShareHttpHealthcheckResponse{
			Healthy: false,
			Error:   err.Error(),
		}, nil
	}
	return &agentGrpc.ShareHttpHealthcheckResponse{Healthy: true}, nil
}
