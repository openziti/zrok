package agent

import (
	"context"
	"errors"
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/rest_client_zrok/metadata"
	restshare "github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

func (a *Agent) deletePublicShare(ctx context.Context, token string) error {
	return a.deleteShareConfirmingAbsence(ctx, "public share", token)
}

// deleteShareConfirmingAbsence deletes a share. a typed not-found from the controller counts as deleted only once
// the account overview confirms the token is gone; the controller answers not-found for more than absence.
func (a *Agent) deleteShareConfirmingAbsence(ctx context.Context, what, token string) error {
	err := sdk.DeleteShareContext(ctx, a.root, &sdk.Share{Token: token})
	var missing *restshare.UnshareNotFound
	if !errors.As(err, &missing) {
		return err
	}
	listed := func(resources *rest_model_zrok.EnvironmentAndResources) bool {
		for _, shr := range resources.Shares {
			if shr != nil && shr.ShareToken == token {
				return true
			}
		}
		return false
	}
	if absenceErr := a.confirmAbsence(ctx, what, token, listed); absenceErr != nil {
		return errors.Join(err, absenceErr)
	}
	return nil
}

// confirmAbsence validates both environment visibility and absence across the authenticated account. listed
// reports whether one environment's resources still hold the token.
func (a *Agent) confirmAbsence(ctx context.Context, what, token string, listed func(*rest_model_zrok.EnvironmentAndResources) bool) error {
	client, err := env_core.ClientContext(ctx, a.root)
	if err != nil {
		return fmt.Errorf("unable to confirm %v absence: %w", what, err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", a.root.Environment().AccountToken)
	response, err := client.Metadata.Overview(metadata.NewOverviewParamsWithContext(ctx), auth)
	if err != nil {
		return fmt.Errorf("unable to confirm %v absence: %w", what, err)
	}
	if response == nil || response.Payload == nil {
		return fmt.Errorf("cannot confirm %v absence without an overview", what)
	}
	foundEnvironment := false
	for _, resources := range response.Payload.Environments {
		if resources == nil {
			continue
		}
		if resources.Environment != nil && resources.Environment.ZID == a.root.Environment().ZitiIdentity {
			foundEnvironment = true
		}
		if listed(resources) {
			return fmt.Errorf("%v '%v' is still listed in account overview", what, token)
		}
	}
	if !foundEnvironment {
		return fmt.Errorf("cannot confirm %v absence: environment '%v' is missing from overview", what, a.root.Environment().ZitiIdentity)
	}
	return nil
}
