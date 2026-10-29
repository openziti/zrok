package agent

import (
	"context"
	"slices"
	"sort"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *agentGrpcImpl) Status(_ context.Context, _ *agentGrpc.StatusRequest) (*agentGrpc.StatusResponse, error) {
	a := i.agent
	response := &agentGrpc.StatusResponse{}
	a.mu.Lock()
	for _, op := range a.operations {
		if op.phase == phaseIdle {
			continue
		}
		status := string(op.phase)
		var failure *agentGrpc.Failure
		if op.failure != nil {
			failure = &agentGrpc.Failure{Id: op.failureID, Count: int32(op.failure.Count), LastError: op.failure.LastError, NextRetry: timestamppb.New(op.failure.NextRetry)}
			if a.cfg.MaxRetries >= 0 && op.failure.Count >= a.cfg.MaxRetries && op.phase == phaseRetry {
				status = "failed"
			}
		} else if op.phase == phaseQueued || op.phase == phaseStarting {
			failure = &agentGrpc.Failure{Id: op.id}
		}
		switch req := op.request.(type) {
		case *AccessPrivateRequest:
			detail := &agentGrpc.AccessDetail{Token: req.ShareToken, BindAddress: req.BindAddress, ResponseHeaders: slices.Clone(req.ResponseHeaders), Status: status, Failure: failure}
			if op.acc != nil {
				detail.FrontendToken = op.acc.frontendToken
				detail.BindAddress = op.acc.bindAddress
			}
			response.Accesses = append(response.Accesses, detail)
		case *SharePublicRequest:
			detail := &agentGrpc.ShareDetail{ShareMode: "public", BackendMode: req.BackendMode, BackendEndpoint: req.Target, Closed: req.Closed, Status: status, Failure: failure}
			if op.shr != nil {
				detail.Token = op.shr.token
				detail.BackendMode = string(op.shr.backendMode)
				detail.BackendEndpoint = op.shr.target
				detail.FrontendEndpoint = slices.Clone(op.shr.frontendEndpoints)
			}
			response.Shares = append(response.Shares, detail)
		case *SharePrivateRequest:
			detail := &agentGrpc.ShareDetail{ShareMode: "private", BackendMode: req.BackendMode, BackendEndpoint: req.Target, Closed: req.Closed, Status: status, Failure: failure}
			if op.shr != nil {
				detail.Token = op.shr.token
				detail.BackendMode = string(op.shr.backendMode)
				detail.BackendEndpoint = op.shr.target
				detail.FrontendEndpoint = slices.Clone(op.shr.frontendEndpoints)
			}
			response.Shares = append(response.Shares, detail)
		}
	}
	a.mu.Unlock()
	sort.Slice(response.Accesses, func(i, j int) bool { return response.Accesses[i].Token < response.Accesses[j].Token })
	sort.Slice(response.Shares, func(i, j int) bool {
		left, right := response.Shares[i], response.Shares[j]
		if left.BackendEndpoint != right.BackendEndpoint {
			return left.BackendEndpoint < right.BackendEndpoint
		}
		return left.Token < right.Token
	})
	return response, nil
}
