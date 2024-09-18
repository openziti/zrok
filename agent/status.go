package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
)

func (i *agentGrpcImpl) Status(_ context.Context, _ *agentGrpc.StatusRequest) (*agentGrpc.StatusResponse, error) {
	var accesses []*agentGrpc.AccessDetail
	for feToken, acc := range i.a.accesses {
		accesses = append(accesses, &agentGrpc.AccessDetail{
			FrontendToken:   feToken,
			Token:           acc.token,
			BindAddress:     acc.bindAddress,
			ResponseHeaders: acc.responseHeaders,
		})
	}

	var shares []*agentGrpc.ShareDetail
	for token, shr := range i.a.shares {
		shares = append(shares, &agentGrpc.ShareDetail{
			Token:            token,
			ShareMode:        string(shr.shareMode),
			BackendMode:      string(shr.backendMode),
			Reserved:         shr.reserved,
			FrontendEndpoint: shr.frontendSelection,
			BackendEndpoint:  shr.target,
			Closed:           shr.closed,
		})
	}

	return &agentGrpc.StatusResponse{Accesses: accesses, Shares: shares}, nil
}
