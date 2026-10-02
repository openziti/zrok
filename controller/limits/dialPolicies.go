package limits

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/pkg/errors"
)

// desiredDialPolicies returns every dial service policy a live share should have: the frontend policy of a
// public share and the policy of each private access to a share of either mode. the errors are store errors.
func desiredDialPolicies(str *store.Store, shr *store.Share, trx *sqlx.Tx) ([]*automation.ServicePolicyOptions, error) {
	var desired []*automation.ServicePolicyOptions
	if shr.ShareMode == string(sdk.PublicShareMode) {
		public, err := desiredPublicDialPolicies(str, shr, trx)
		if err != nil {
			return nil, err
		}
		desired = append(desired, public...)
	}
	access, err := desiredAccessDialPolicies(str, shr, trx)
	if err != nil {
		return nil, err
	}
	return append(desired, access...), nil
}

// desiredPublicDialPolicies derives a public share's frontend dial policy the way allocatePublicResources
// creates it: from the share's live name mappings, each name's namespace, and that namespace's frontends.
// frontend_selection is v1's column and v2 never writes it, so it is consulted only for a v1-created row,
// one with no live mappings. a share with neither has no dial policy to restore; its create made none.
func desiredPublicDialPolicies(str *store.Store, shr *store.Share, trx *sqlx.Tx) ([]*automation.ServicePolicyOptions, error) {
	env, err := str.GetEnvironment(shr.EnvironmentId, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding environment for share '%v'", shr.Token)
	}
	details, err := str.FindShareNameCleanupDetailsByShareId(shr.Id, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding name mappings for share '%v'", shr.Token)
	}

	var frontendZIds []string
	mapped := false
	seenNamespaces := make(map[int]bool)
	seenFrontends := make(map[string]bool)
	for _, detail := range details {
		if detail.NameDeleted || detail.NamespaceDeleted {
			continue
		}
		mapped = true
		if seenNamespaces[detail.NamespaceID] {
			continue
		}
		seenNamespaces[detail.NamespaceID] = true
		fes, err := str.FindFrontendsForNamespace(detail.NamespaceID, trx)
		if err != nil {
			return nil, errors.Wrapf(err, "error finding frontends for namespace '%v'", detail.NamespaceName)
		}
		for _, fe := range fes {
			if !seenFrontends[fe.ZId] {
				seenFrontends[fe.ZId] = true
				frontendZIds = append(frontendZIds, fe.ZId)
			}
		}
	}

	if !mapped && shr.FrontendSelection != nil {
		fe, err := str.FindFrontendPubliclyNamed(*shr.FrontendSelection, trx)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, missingPublicFrontendError{shareToken: shr.Token, publicName: *shr.FrontendSelection}
		}
		if err != nil {
			return nil, errors.Wrapf(err, "error finding frontend name '%v' for '%v'", *shr.FrontendSelection, shr.Token)
		}
		frontendZIds = []string{fe.ZId}
	}

	if len(frontendZIds) == 0 {
		return nil, nil
	}
	var identityRoles []string
	for _, zId := range frontendZIds {
		identityRoles = append(identityRoles, "@"+zId)
	}
	return []*automation.ServicePolicyOptions{{
		BaseOptions: automation.BaseOptions{
			Name: env.ZId + "-" + shr.ZId + "-dial",
			Tags: automation.ZrokShareTags(shr.Token),
		},
		IdentityRoles: identityRoles,
		ServiceRoles:  []string{"@" + shr.ZId},
		PolicyType:    rest_model.DialBindDial,
		Semantic:      rest_model.SemanticAllOf,
	}}, nil
}

// missingPublicFrontendError is a v1-created share whose frontend selection names no live frontend. it is
// the share's failure, not the store's: the lookup found no row, which aborts no transaction.
type missingPublicFrontendError struct {
	shareToken, publicName string
}

func (e missingPublicFrontendError) Error() string {
	return fmt.Sprintf("share '%v' selects frontend '%v', which does not exist", e.shareToken, e.publicName)
}

// desiredAccessDialPolicies returns the dial policy of each private access to the share, as access creates
// it. access does not check the share mode, so a public share can have these too.
func desiredAccessDialPolicies(str *store.Store, shr *store.Share, trx *sqlx.Tx) ([]*automation.ServicePolicyOptions, error) {
	fes, err := str.FindFrontendsForPrivateShare(shr.Id, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding frontends for share '%v'", shr.Token)
	}
	var desired []*automation.ServicePolicyOptions
	for _, fe := range fes {
		if fe.EnvironmentId == nil {
			continue
		}
		env, err := str.GetEnvironment(*fe.EnvironmentId, trx)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting environment for frontend '%v'", fe.Token)
		}
		desired = append(desired, &automation.ServicePolicyOptions{
			BaseOptions: automation.BaseOptions{
				Name: fe.Token + "-" + env.ZId + "-" + shr.ZId + "-dial",
				Tags: automation.ZrokShareTags(shr.Token).
					WithTag("zrokEnvironmentZId", env.ZId).
					WithTag("zrokFrontendToken", fe.Token),
			},
			IdentityRoles: []string{"@" + env.ZId},
			ServiceRoles:  []string{"@" + shr.ZId},
			PolicyType:    rest_model.DialBindDial,
			Semantic:      rest_model.SemanticAllOf,
		})
	}
	return desired, nil
}

// dialPolicyExists looks the policy up by its deterministic name.
func dialPolicyExists(ziti *automation.ZitiAutomation, name string) (bool, error) {
	policies, err := ziti.ServicePolicies.Find(&automation.FilterOptions{Filter: automation.BuildFilter("name", name)})
	if err != nil {
		return false, errors.Wrapf(err, "error finding dial service policy '%v'", name)
	}
	return len(policies) > 0, nil
}

// ensureDialPolicy creates the policy unless one of its name exists, and reports whether it created it.
func ensureDialPolicy(ziti *automation.ZitiAutomation, opts *automation.ServicePolicyOptions) (bool, error) {
	exists, err := dialPolicyExists(ziti, opts.Name)
	if err != nil {
		return false, err
	}
	if exists {
		dl.Debugf("dial service policy '%v' already exists", opts.Name)
		return false, nil
	}
	if _, err := ziti.ServicePolicies.CreateDial(opts); err != nil {
		return false, errors.Wrapf(err, "error creating dial service policy '%v'", opts.Name)
	}
	return true, nil
}
