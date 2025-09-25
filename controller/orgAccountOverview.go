package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
)

type orgAccountOverviewHandler struct{}

func newOrgAccountOverviewHandler() *orgAccountOverviewHandler {
	return &orgAccountOverviewHandler{}
}

func (h *orgAccountOverviewHandler) Handle(params metadata.OrgAccountOverviewParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return metadata.NewOrgAccountOverviewInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	org, err := str.FindOrganizationByToken(params.OrganizationToken, trx)
	if err != nil {
		dl.Errorf("error finding organization by token: %v", err)
		return metadata.NewOrgAccountOverviewNotFound()
	}

	admin, err := str.IsAccountAdminOfOrganization(int(principal.ID), org.Id, trx)
	if err != nil {
		dl.Errorf("error checking account '%v' admin: %v", principal.Email, err)
		return metadata.NewOrgAccountOverviewNotFound()
	}
	if !admin {
		dl.Errorf("requesting account '%v' is not admin of organization '%v'", principal.Email, org.Token)
		return metadata.NewOrgAccountOverviewNotFound()
	}

	acct, err := str.FindAccountWithEmail(params.AccountEmail, trx)
	if err != nil {
		dl.Errorf("error finding account by email: %v", err)
		return metadata.NewOrgAccountOverviewNotFound()
	}

	inOrg, err := str.IsAccountInOrganization(acct.Id, org.Id, trx)
	if err != nil {
		dl.Errorf("error checking account '%v' organization membership: %v", acct.Email, err)
		return metadata.NewOrgAccountOverviewNotFound()
	}
	if !inOrg {
		dl.Errorf("account '%v' is not a member of organization '%v'", acct.Email, org.Token)
		return metadata.NewOrgAccountOverviewNotFound()
	}

	envs, err := str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		dl.Errorf("error finding environments for '%v': %v", acct.Email, err)
		return metadata.NewOrgAccountOverviewNotFound()
	}

	accountLimited, err := isAccountLimited(acct.Id, trx)
	if err != nil {
		dl.Errorf("error checking account '%v' limited: %v", acct.Email, err)
	}

	ovr := &rest_model_zrok.Overview{AccountLimited: accountLimited}

	for _, env := range envs {
		ear := &rest_model_zrok.EnvironmentAndResources{
			Environment: &rest_model_zrok.Environment{
				Address:     env.Address,
				Description: env.Description,
				Host:        env.Host,
				ZID:         env.ZId,
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
			if fe.PrivateShareId != nil {
				feShr, err := str.GetShare(*fe.PrivateShareId, trx)
				if err != nil {
					dl.Errorf("error getting share for frontend '%v': %v", fe.ZId, err)
					return metadata.NewOverviewInternalServerError()
				}
				envFe.ShareToken = feShr.Token
			}
			ear.Frontends = append(ear.Frontends, envFe)
		}

		ovr.Environments = append(ovr.Environments, ear)
	}

	return metadata.NewOrgAccountOverviewOK().WithPayload(ovr)
}
