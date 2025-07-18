package agent

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/sdk/golang/sdk"
)

func (a *Agent) HttpShareHealthcheck(shareToken, endpoint, httpVerb, expectedHttpResponse string, timeoutMs int) error {
	if shr, found := a.shares[shareToken]; found {
		if shr.backendMode == sdk.ProxyBackendMode {
			return a.doHealthcheckRequest(shr, endpoint, httpVerb, expectedHttpResponse, timeoutMs)
		} else {
			return fmt.Errorf("cannot perform http healthcheck on '%v' share '%v'", shr.backendMode, shareToken)
		}
	} else {
		return fmt.Errorf("share '%v' not found in agent", shareToken)
	}
}

func (a *Agent) doHealthcheckRequest(shr *share, endpoint, httpVerb, expectedHttpResponse string, timeoutMs int) error {
	expectedStatusCode, err := strconv.Atoi(expectedHttpResponse)
	if err != nil {
		return fmt.Errorf("provided expected http status '%v' is invalid: %v", expectedHttpResponse, err)
	}
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
	if resp.StatusCode != expectedStatusCode {
		return fmt.Errorf("unexpected status code; got '%v', want '%v'", resp.StatusCode, expectedStatusCode)
	}
	return nil
}

func (i *agentGrpcImpl) HttpShareHealthcheck(_ context.Context, req *agentGrpc.HttpShareHealthcheckRequest) (*agentGrpc.HttpShareHealthcheckResponse, error) {
	if err := i.agent.HttpShareHealthcheck(req.Token, req.Endpoint, req.HttpVerb, req.ExpectedHttpResponse, int(req.TimeoutMs)); err != nil {
		return &agentGrpc.HttpShareHealthcheckResponse{
			Healthy: false,
			Error:   err.Error(),
		}, nil
	}
	return &agentGrpc.HttpShareHealthcheckResponse{Healthy: true}, nil
}
