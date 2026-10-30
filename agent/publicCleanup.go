package agent

import (
	"context"
	"errors"
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/rest_client_zrok/metadata"
	restshare "github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

func (a *Agent) deletePublicShare(ctx context.Context, token string) error {
	err := sdk.DeleteShareContext(ctx, a.root, &sdk.Share{Token: token})
	var missing *restshare.UnshareNotFound
	if !errors.As(err, &missing) {
		return err
	}
	if absenceErr := a.confirmPublicAbsence(ctx, token); absenceErr != nil {
		return errors.Join(err, absenceErr)
	}
	return nil
}

// confirmPublicAbsence validates both environment visibility and absence across the authenticated account.
func (a *Agent) confirmPublicAbsence(ctx context.Context, token string) error {
	client, err := env_core.ClientContext(ctx, a.root)
	if err != nil {
		return fmt.Errorf("unable to confirm public share absence: %w", err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", a.root.Environment().AccountToken)
	response, err := client.Metadata.Overview(metadata.NewOverviewParamsWithContext(ctx), auth)
	if err != nil {
		return fmt.Errorf("unable to confirm public share absence: %w", err)
	}
	if response == nil || response.Payload == nil {
		return errors.New("cannot confirm public share absence without an overview")
	}
	foundEnvironment := false
	for _, resources := range response.Payload.Environments {
		if resources == nil {
			continue
		}
		if resources.Environment != nil && resources.Environment.ZID == a.root.Environment().ZitiIdentity {
			foundEnvironment = true
		}
		for _, shr := range resources.Shares {
			if shr != nil && shr.ShareToken == token {
				return fmt.Errorf("public share '%v' is still listed in account overview", token)
			}
		}
	}
	if !foundEnvironment {
		return fmt.Errorf("cannot confirm public share absence: environment '%v' is missing from overview", a.root.Environment().ZitiIdentity)
	}
	return nil
}
