package agent

import (
	"context"
	"sort"

	"github.com/openziti/zrok/agent/agentGrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *agentGrpcImpl) Status(_ context.Context, _ *agentGrpc.StatusRequest) (*agentGrpc.StatusResponse, error) {
	var accesses []*agentGrpc.AccessDetail

	// active accesses
	for feToken, acc := range i.agent.accesses {
		accesses = append(accesses, &agentGrpc.AccessDetail{
			FrontendToken:   feToken,
			Token:           acc.token,
			BindAddress:     acc.bindAddress,
			ResponseHeaders: acc.responseHeaders,
			Status:          "active",
		})
	}

	// failed accesses
	for failureId, access := range i.agent.retryManager.accesses {
		status := "retrying"
		if i.agent.cfg.MaxRetries > -1 && access.Failure.Count >= i.agent.cfg.MaxRetries {
			status = "failed"
		}
		accesses = append(accesses, &agentGrpc.AccessDetail{
			FrontendToken:   "",
			Token:           access.Request.Token,
			BindAddress:     access.Request.BindAddress,
			ResponseHeaders: access.Request.ResponseHeaders,
			Status:          status,
			Failure: &agentGrpc.Failure{
				Id:        failureId,
				Count:     int32(access.Failure.Count),
				LastError: access.Failure.LastError,
				NextRetry: timestamppb.New(access.Failure.NextRetry),
			},
		})
	}

	sort.Slice(accesses, func(i, j int) bool {
		return accesses[i].Token < accesses[j].Token
	})

	var shares []*agentGrpc.ShareDetail

	// active shares
	for token, shr := range i.agent.shares {
		shares = append(shares, &agentGrpc.ShareDetail{
			Token:            token,
			ShareMode:        string(shr.shareMode),
			BackendMode:      string(shr.backendMode),
			FrontendEndpoint: shr.frontendEndpoints,
			BackendEndpoint:  shr.target,
			Closed:           shr.closed,
			Status:           "active",
		})
	}

	// Add failed shares with failure IDs
	for failureId, share := range i.agent.retryManager.shares {
		status := "retrying"
		if i.agent.cfg.MaxRetries > -1 && share.Failure.Count >= i.agent.cfg.MaxRetries {
			status = "failed"
		}
		shares = append(shares, &agentGrpc.ShareDetail{
			Token:            "",
			ShareMode:        "public",
			BackendMode:      share.Request.BackendMode,
			FrontendEndpoint: nil,
			BackendEndpoint:  share.Request.Target,
			Closed:           share.Request.Closed,
			Status:           status,
			Failure: &agentGrpc.Failure{
				Id:        failureId,
				Count:     int32(share.Failure.Count),
				LastError: share.Failure.LastError,
				NextRetry: timestamppb.New(share.Failure.NextRetry),
			},
		})
	}

	sort.Slice(shares, func(i, j int) bool {
		if shares[i].BackendEndpoint != shares[j].BackendEndpoint {
			return shares[i].BackendEndpoint < shares[j].BackendEndpoint
		}
		return shares[i].Token < shares[j].Token
	})

	return &agentGrpc.StatusResponse{Accesses: accesses, Shares: shares}, nil
}
