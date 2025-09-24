package agent

import (
	"context"
	"sort"

	"github.com/openziti/zrok/agent/agentGrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *agentGrpcImpl) Status(_ context.Context, _ *agentGrpc.StatusRequest) (*agentGrpc.StatusResponse, error) {
	var accesses []*agentGrpc.AccessDetail

	// Add active accesses
	for feToken, acc := range i.agent.accesses {
		accesses = append(accesses, &agentGrpc.AccessDetail{
			FrontendToken:   feToken,
			Token:           acc.token,
			BindAddress:     acc.bindAddress,
			ResponseHeaders: acc.responseHeaders,
			Status:          "active",
			FailureId:       "",
			FailureCount:    0,
			LastError:       "",
		})
	}

	// Add failed accesses with failure IDs
	for failureID, entry := range i.agent.failedAccesses {
		var lastFailure, nextRetry *timestamppb.Timestamp
		if entry.LastFailure != nil {
			lastFailure = timestamppb.New(*entry.LastFailure)
		}
		if entry.NextRetry != nil {
			nextRetry = timestamppb.New(*entry.NextRetry)
		}

		status := "retrying"
		if i.agent.cfg.MaxRetries > -1 && entry.FailureCount >= i.agent.cfg.MaxRetries {
			status = "failed"
		}

		accesses = append(accesses, &agentGrpc.AccessDetail{
			FrontendToken:   "",
			Token:           entry.Request.Token,
			BindAddress:     entry.Request.BindAddress,
			ResponseHeaders: entry.Request.ResponseHeaders,
			Status:          status,
			FailureId:       failureID,
			FailureCount:    int32(entry.FailureCount),
			LastError:       entry.LastError,
			LastFailure:     lastFailure,
			NextRetry:       nextRetry,
		})
	}

	sort.Slice(accesses, func(i, j int) bool {
		return accesses[i].Token < accesses[j].Token
	})

	var shares []*agentGrpc.ShareDetail

	// Add active shares
	for token, shr := range i.agent.shares {
		shares = append(shares, &agentGrpc.ShareDetail{
			Token:            token,
			ShareMode:        string(shr.shareMode),
			BackendMode:      string(shr.backendMode),
			FrontendEndpoint: shr.frontendEndpoints,
			BackendEndpoint:  shr.target,
			Closed:           shr.closed,
			Status:           "active",
			FailureId:        "",
			FailureCount:     0,
			LastError:        "",
		})
	}

	// Add failed shares with failure IDs
	for failureID, entry := range i.agent.failedShares {
		var lastFailure, nextRetry *timestamppb.Timestamp
		if entry.LastFailure != nil {
			lastFailure = timestamppb.New(*entry.LastFailure)
		}
		if entry.NextRetry != nil {
			nextRetry = timestamppb.New(*entry.NextRetry)
		}

		status := "retrying"
		if i.agent.cfg.MaxRetries > -1 && entry.FailureCount >= i.agent.cfg.MaxRetries {
			status = "failed"
		}

		shares = append(shares, &agentGrpc.ShareDetail{
			Token:            "",
			ShareMode:        "public",
			BackendMode:      entry.Request.BackendMode,
			FrontendEndpoint: []string{},
			BackendEndpoint:  entry.Request.Target,
			Closed:           entry.Request.Closed,
			Status:           status,
			FailureId:        failureID,
			FailureCount:     int32(entry.FailureCount),
			LastError:        entry.LastError,
			LastFailure:      lastFailure,
			NextRetry:        nextRetry,
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
