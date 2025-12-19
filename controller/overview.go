package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/metadata"
)

type overviewHandler struct{}

func newOverviewHandler() *overviewHandler {
	return &overviewHandler{}
}

func (h *overviewHandler) Handle(_ metadata.OverviewParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return metadata.NewOverviewInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environments for '%v': %v", principal.Email, err)
		return metadata.NewOverviewInternalServerError()
	}

	accountLimited, err := isAccountLimited(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error checking account limited for '%v': %v", principal.Email, err)
	}

	ovr := &rest_model_zrok.Overview{AccountLimited: accountLimited}
	for _, env := range envs {
		remoteAgent, err := str.IsAgentEnrolledForEnvironment(env.Id, trx)
		if err != nil {
			dl.Errorf("error checking agent enrollment for environment '%v' (%v): %v", env.ZId, principal.Email, err)
			return metadata.NewOverviewInternalServerError()
		}

		ear := &rest_model_zrok.EnvironmentAndResources{
			Environment: &rest_model_zrok.Environment{
				Address:     env.Address,
				Description: env.Description,
				Host:        env.Host,
				ZID:         env.ZId,
				RemoteAgent: remoteAgent,
				Limited:     accountLimited,
				CreatedAt:   env.CreatedAt.UnixMilli(),
				UpdatedAt:   env.UpdatedAt.UnixMilli(),
			},
		}

		shrs, err := str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			dl.Errorf("error finding shares for environment '%v': %v", env.ZId, err)
			return metadata.NewOverviewInternalServerError()
		}
		for _, shr := range shrs {
			frontendEndpoints := buildFrontendEndpointsForShare(shr.Id, shr.Token, shr.FrontendEndpoint, trx)
			target := ""
			if shr.BackendProxyEndpoint != nil {
				target = *shr.BackendProxyEndpoint
			}
			envShr := &rest_model_zrok.Share{
				ShareToken:        shr.Token,
				ZID:               shr.ZId,
				EnvZID:            env.ZId,
				ShareMode:         shr.ShareMode,
				BackendMode:       shr.BackendMode,
				FrontendEndpoints: frontendEndpoints,
				Target:            target,
				Limited:           accountLimited,
				CreatedAt:         shr.CreatedAt.UnixMilli(),
				UpdatedAt:         shr.UpdatedAt.UnixMilli(),
			}
			ear.Shares = append(ear.Shares, envShr)
		}
		fes, err := str.FindFrontendsForEnvironment(env.Id, trx)
		if err != nil {
			dl.Errorf("error finding frontends for environment '%v': %v", env.ZId, err)
			return metadata.NewOverviewInternalServerError()
		}
		for _, fe := range fes {
			envFe := &rest_model_zrok.Frontend{
				ID:            int64(fe.Id),
				FrontendToken: fe.Token,
				ZID:           fe.ZId,
				CreatedAt:     fe.CreatedAt.UnixMilli(),
				UpdatedAt:     fe.UpdatedAt.UnixMilli(),
			}
			if fe.BindAddress != nil {
				envFe.BindAddress = *fe.BindAddress
			}
			if fe.Description != nil {
				envFe.Description = *fe.Description
			}
			if fe.PrivateShareId != nil {
				feShr, err := str.GetShare(*fe.PrivateShareId, trx)
				if err != nil {
					dl.Errorf("error getting share for frontend '%v': %v", fe.ZId, err)
					return metadata.NewOverviewInternalServerError()
				}
				envFe.ShareToken = feShr.Token
				envFe.BackendMode = feShr.BackendMode
			}
			ear.Frontends = append(ear.Frontends, envFe)
		}

		ovr.Environments = append(ovr.Environments, ear)
	}

	// find all namespaces the user has access to
	namespaces, err := str.FindNamespacesForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding namespaces for account '%v': %v", principal.Email, err)
		return metadata.NewOverviewInternalServerError()
	}

	// build namespace objects for overview
	for _, namespace := range namespaces {
		ovr.Namespaces = append(ovr.Namespaces, &rest_model_zrok.OverviewNamespacesItems0{
			NamespaceToken: namespace.Token,
			Name:           namespace.Name,
			Description:    namespace.Description,
		})
	}

	// collect allocated names from all accessible namespaces
	for _, ns := range namespaces {
		names, err := str.FindNamesWithShareTokensForAccountAndNamespace(int(principal.ID), ns.Id, trx)
		if err != nil {
			dl.Errorf("error finding allocated names for namespace '%v': %v", ns.Token, err)
			return metadata.NewOverviewInternalServerError()
		}

		for _, an := range names {
			nameObj := &rest_model_zrok.OverviewNamesItems0{
				NamespaceToken: ns.Token,
				NamespaceName:  ns.Name,
				Name:           an.Name.Name,
				Reserved:       an.Name.Reserved,
				CreatedAt:      an.Name.CreatedAt.Unix(),
			}
			if an.ShareToken != nil {
				nameObj.ShareToken = *an.ShareToken
			}
			ovr.Names = append(ovr.Names, nameObj)
		}
	}

	return metadata.NewOverviewOK().WithPayload(ovr)
}
