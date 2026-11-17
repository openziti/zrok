package agent

import (
	"context"
	"errors"

	restshare "github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

// private shares and accesses keep their legacy creation ownership: their workers create them. the agent deletes
// them, and a resource the controller has already removed must not leave its operation unable to finish cleanup.
// the controller removes every access to a share when that share is torn down, so an access is routinely gone
// before the agent asks.

func (a *Agent) deletePrivateShare(ctx context.Context, token string) error {
	return a.deleteShareConfirmingAbsence(ctx, "private share", token)
}

// deleteAccess deletes an access by its frontend token. a typed not-found counts as deleted only once the account
// overview confirms no environment still lists that frontend.
func (a *Agent) deleteAccess(ctx context.Context, shareToken, frontendToken string) error {
	err := sdk.DeleteAccessContext(ctx, a.root, &sdk.Access{Token: frontendToken, ShareToken: shareToken})
	var missing *restshare.UnaccessNotFound
	if !errors.As(err, &missing) {
		return err
	}
	listed := func(resources *rest_model_zrok.EnvironmentAndResources) bool {
		for _, frontend := range resources.Frontends {
			if frontend != nil && frontend.FrontendToken == frontendToken {
				return true
			}
		}
		return false
	}
	if absenceErr := a.confirmAbsence(ctx, "access", frontendToken, listed); absenceErr != nil {
		return errors.Join(err, absenceErr)
	}
	return nil
}
