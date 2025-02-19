package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"sort"
)

func (i *agentGrpcImpl) Status(_ context.Context, _ *agentGrpc.StatusRequest) (*agentGrpc.StatusResponse, error) {
	var accesses []*agentGrpc.AccessDetail
	for feToken, acc := range i.agent.accesses {
		accesses = append(accesses, &agentGrpc.AccessDetail{
			FrontendToken:   feToken,
			Token:           acc.token,
			BindAddress:     acc.bindAddress,
			ResponseHeaders: acc.responseHeaders,
		})
	}
	sort.Slice(accesses, func(i, j int) bool {
		return accesses[i].FrontendToken < accesses[j].FrontendToken
	})

	var shares []*agentGrpc.ShareDetail
	for token, shr := range i.agent.shares {
		shares = append(shares, &agentGrpc.ShareDetail{
			Token:            token,
			ShareMode:        string(shr.shareMode),
			BackendMode:      string(shr.backendMode),
			Reserved:         shr.reserved,
			FrontendEndpoint: shr.frontendEndpoints,
			BackendEndpoint:  shr.target,
			Closed:           shr.closed,
		})
	}
	sort.Slice(shares, func(i, j int) bool {
		return shares[i].Token < shares[j].Token
	})

	return &agentGrpc.StatusResponse{Accesses: accesses, Shares: shares}, nil
}
