package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type share12Handler struct{}

func newShare12Handler() *share12Handler {
	return &share12Handler{}
}

func (h *share12Handler) Handle(params share.Share12Params, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewShare12InternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// validate environment
	envZId := params.Body.EnvZID
	envId, err := h.validateEnvironment(envZId, principal, trx)
	if err != nil {
		logrus.Errorf("environment validation failed: %v", err)
		return share.NewShare12Unauthorized()
	}

	// check limits
	if err := h.checkLimits(envId, principal, params, trx); err != nil {
		logrus.Errorf("limits error: %v", err)
		return share.NewShare12Unauthorized()
	}

	// create share token
	shrToken, err := createShareToken()
	if err != nil {
		logrus.Error(err)
		return share.NewShare12InternalServerError()
	}

	// process namespace selections
	frontendEndpoints, nameIds, err := h.processNamespaceSelections(params.Body.NamespaceSelections, shrToken, principal, trx)
	if err != nil {
		logrus.Errorf("namespace selection processing failed: %v", err)
		return share.NewShare12Conflict().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	// allocate resources based on share mode
	var shrZId string
	switch params.Body.ShareMode {
	case "public":
		interstitial, err := h.shouldUseInterstitial(params.Body.BackendMode, principal, trx)
		if err != nil {
			logrus.Errorf("error determining interstitial setting for account '%v': %v", principal.Email, err)
			return share.NewShare12InternalServerError()
		}
		shrZId, frontendEndpoints, err = h.allocatePublicResources(envZId, shrToken, frontendEndpoints, params, interstitial, trx)
	case "private":
		shrZId, frontendEndpoints, err = h.allocatePrivateResources(envZId, shrToken, frontendEndpoints, params, trx)
	default:
		logrus.Errorf("unknown share mode '%v'", params.Body.ShareMode)
		return share.NewShare12InternalServerError()
	}
	if err != nil {
		logrus.Errorf("error allocating share resources: %v", err)
		return share.NewShare12InternalServerError()
	}

	// create share record
	shareId, err := h.createShareRecord(envId, shrZId, shrToken, params, frontendEndpoints, trx)
	if err != nil {
		logrus.Errorf("error creating share record: %v", err)
		return share.NewShare12InternalServerError()
	}

	// create share name mappings for namespace selections
	for _, nameId := range nameIds {
		snm := &store.ShareNameMapping{
			ShareId: shareId,
			NameId:  nameId,
		}
		_, err := str.CreateShareNameMapping(snm, trx)
		if err != nil {
			logrus.Errorf("error creating share name mapping for share '%v' and name '%v': %v", shareId, nameId, err)
			return share.NewShare12InternalServerError()
		}
	}

	// send mapping updates to dynamic frontends after successful commit
	if err := h.processDynamicMappings(shrToken, nameIds, trx); err != nil {
		logrus.Errorf("error sending mapping updates: %v", err)
	}

	// handle access grants if closed permission mode
	if err := h.processAccessGrants(shareId, params.Body.AccessGrants, params.Body.PermissionMode, principal, trx); err != nil {
		logrus.Errorf("error processing access grants: %v", err)
		return share.NewShare12InternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing share record: %v", err)
		return share.NewShare12InternalServerError()
	}

	logrus.Infof("recorded share '%v' with id '%v' for '%v'", shrToken, shareId, principal.Email)

	return share.NewShare12Created().WithPayload(&rest_model_zrok.ShareResponse{
		FrontendProxyEndpoints: frontendEndpoints,
		ShareToken:             shrToken,
	})
}

func (h *share12Handler) validateEnvironment(envZId string, principal *rest_model_zrok.Principal, trx *sqlx.Tx) (int, error) {
	env, err := str.FindEnvironmentForAccount(envZId, int(principal.ID), trx)
	if err != nil {
		return 0, errors.Wrapf(err, "error finding environment '%v' for account '%v'", envZId, principal.Email)
	}
	return env.Id, nil
}

func (h *share12Handler) checkLimits(envId int, principal *rest_model_zrok.Principal, params share.Share12Params, trx *sqlx.Tx) error {
	if !principal.Limitless {
		if limitsAgent != nil {
			shareMode := sdk.ShareMode(params.Body.ShareMode)
			backendMode := sdk.BackendMode(params.Body.BackendMode)

			// we're going to skip reservation checking because we're moving name creation outside the scope of share
			// creation. the limits check for name creation will happen in the `/share/name` endpoint instead.
			ok, err := limitsAgent.CanCreateShare(int(principal.ID), envId, false, false, shareMode, backendMode, trx)
			if err != nil {
				return errors.Wrapf(err, "error checking share limits for '%v'", principal.Email)
			}
			if !ok {
				return errors.Errorf("share limit check failed for '%v'", principal.Email)
			}
		}
	}
	return nil
}

func (h *share12Handler) shouldUseInterstitial(backendMode string, principal *rest_model_zrok.Principal, trx *sqlx.Tx) (bool, error) {
	var skipInterstitial bool
	parsedBackendMode := sdk.BackendMode(backendMode)

	if parsedBackendMode != sdk.DriveBackendMode {
		var err error
		skipInterstitial, err = str.IsAccountGrantedSkipInterstitial(int(principal.ID), trx)
		if err != nil {
			return false, errors.Wrapf(err, "error checking skip interstitial for account '%v'", principal.Email)
		}
	} else {
		// always skip interstitial for drive backend mode
		skipInterstitial = true
	}

	return !skipInterstitial, nil
}

func (h *share12Handler) processNamespaceSelections(selections []*rest_model_zrok.NamespaceSelection, shrToken string, principal *rest_model_zrok.Principal, trx *sqlx.Tx) ([]string, []int, error) {
	var frontendEndpoints []string
	var nameIds []int

	for _, selection := range selections {
		// find namespace by token
		ns, err := str.FindNamespaceWithToken(selection.NamespaceToken, trx)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error finding namespace with token '%v'", selection.NamespaceToken)
		}

		var endpoint string
		var nameId int

		if selection.Name != "" { // user specified a name - validate ownership and availability
			name, err := str.FindNameByNamespaceAndName(ns.Id, selection.Name, trx)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error finding name '%v' in namespace '%v'", selection.Name, ns.Token)
			}

			// check if user owns this name
			if name.AccountId != int(principal.ID) {
				return nil, nil, errors.Errorf("user '%v' does not own name '%v' in namespace '%v'", principal.Email, selection.Name, ns.Token)
			}

			// check if there's already a share_name_mapping for this name
			existing, err := str.FindShareNameMappingsByNameId(name.Id, trx)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error checking existing share name mappings for name '%v'", selection.Name)
			}
			if len(existing) > 0 {
				return nil, nil, errors.Errorf("name '%v' in namespace '%v' is already in use by another share", selection.Name, ns.Token)
			}

			nameId = name.Id
			endpoint = util.ExpandUrlTemplate(name.Name, ns.Name)

		} else { // no name specified - generate one and create name record
			// check namespace permissions
			if !ns.Open {
				granted, err := str.CheckNamespaceGrant(ns.Id, int(principal.ID), trx)
				if err != nil {
					return nil, nil, errors.Wrapf(err, "error checking namespace grant for account '%v' and namespace '%v'", principal.Email, ns.Token)
				}
				if !granted {
					return nil, nil, errors.Errorf("account '%v' is not granted access to namespace '%v'", principal.Email, ns.Token)
				}
			}

			// create name record with reserved=false (dynamically allocated)
			name := &store.Name{
				NamespaceId: ns.Id,
				Name:        shrToken,
				AccountId:   int(principal.ID),
				Reserved:    false,
			}

			nameId, err = str.CreateName(name, trx)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error creating allocated name '%v' in namespace '%v' for account '%v'", shrToken, ns.Token, principal.Email)
			}

			endpoint = util.ExpandUrlTemplate(shrToken, ns.Name)
		}

		frontendEndpoints = append(frontendEndpoints, endpoint)
		nameIds = append(nameIds, nameId)
	}

	return frontendEndpoints, nameIds, nil
}

func (h *share12Handler) allocatePublicResources(envZId, shrToken string, frontendEndpoints []string, params share.Share12Params, interstitial bool, trx interface{}) (string, []string, error) {
	// create automation client
	automationCfg := &automation.Config{
		ApiEndpoint: cfg.Ziti.ApiEndpoint,
		Username:    cfg.Ziti.Username,
		Password:    cfg.Ziti.Password,
	}
	ziti, err := automation.NewZitiAutomation(automationCfg)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating ziti automation client")
	}

	// prepare auth users
	var authUsers []*sdk.AuthUserConfig
	for _, authUser := range params.Body.BasicAuthUsers {
		authUsers = append(authUsers, &sdk.AuthUserConfig{Username: authUser.Username, Password: authUser.Password})
	}

	// parse auth scheme
	authScheme, err := sdk.ParseAuthScheme(params.Body.AuthScheme)
	if err != nil {
		return "", nil, errors.Wrap(err, "error parsing auth scheme")
	}

	// prepare oauth config
	var oauthCfg *sdk.OauthConfig
	if authScheme == sdk.Oauth {
		oauthCfg = &sdk.OauthConfig{
			Provider:                   params.Body.OauthProvider,
			EmailDomains:               params.Body.OauthEmailDomains,
			AuthorizationCheckInterval: params.Body.OauthRefreshInterval,
		}
	}

	// create frontend config
	frontendConfig := &sdk.FrontendConfig{
		Interstitial: interstitial,
		AuthScheme:   authScheme,
	}
	if authScheme == sdk.Basic {
		frontendConfig.BasicAuth = &sdk.BasicAuthConfig{Users: authUsers}
	}
	if authScheme == sdk.Oauth && oauthCfg != nil {
		frontendConfig.OauthAuth = oauthCfg
	}

	// create config using the global zrokProxyConfigId
	tags := automation.ZrokShareTags(shrToken)
	configOpts := &automation.ConfigOptions{
		BaseOptions: automation.BaseOptions{
			Name: shrToken,
			Tags: tags,
		},
		ConfigTypeID: zrokProxyConfigId,
		Data:         frontendConfig,
	}
	cfgZId, err := ziti.Configs.Create(configOpts)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating config")
	}

	// create share service
	serviceOpts := &automation.ServiceOptions{
		BaseOptions: automation.BaseOptions{
			Name: shrToken,
			Tags: tags,
		},
		Configs:            []string{cfgZId},
		EncryptionRequired: true,
	}
	shrZId, err := ziti.Services.Create(serviceOpts)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating share service")
	}

	// create bind policy (backend can bind to this service)
	bindPolicyName := envZId + "-" + shrZId + "-bind"
	bindPolicy := automation.NewPolicyBuilder(bindPolicyName).
		WithServiceIDs(shrZId).
		WithIdentityIDs(envZId).
		WithTags(tags, nil)
	_, err = ziti.Policies.CreateServicePolicyBind(bindPolicy)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating service policy bind")
	}

	// create dial policy (frontends can dial this service)
	// get frontend identities from namespaces
	var frontendZIds []string
	for _, selection := range params.Body.NamespaceSelections {
		ns, err := str.FindNamespaceWithToken(selection.NamespaceToken, trx.(*sqlx.Tx))
		if err != nil {
			return "", nil, errors.Wrapf(err, "error finding namespace with token '%v'", selection.NamespaceToken)
		}

		frontends, err := str.FindFrontendsForNamespace(ns.Id, trx.(*sqlx.Tx))
		if err != nil {
			return "", nil, errors.Wrapf(err, "error finding frontends for namespace '%v'", ns.Token)
		}

		for _, fe := range frontends {
			frontendZIds = append(frontendZIds, fe.ZId)
		}
	}

	if len(frontendZIds) > 0 {
		dialPolicyName := envZId + "-" + shrZId + "-dial"
		dialPolicy := automation.NewPolicyBuilder(dialPolicyName).
			WithServiceIDs(shrZId).
			WithIdentityIDs(frontendZIds...).
			WithTags(tags, nil)
		_, err = ziti.Policies.CreateServicePolicyDial(dialPolicy)
		if err != nil {
			return "", nil, errors.Wrap(err, "error creating service policy dial")
		}
	}

	// create service edge router policy
	serpPolicyName := envZId + "-" + shrToken + "-serp"
	serpPolicy := automation.NewPolicyBuilder(serpPolicyName).
		WithServiceIDs(shrZId).
		WithAllEdgeRouters().
		WithTags(tags, nil)
	_, err = ziti.Policies.CreateServiceEdgeRouterPolicy(serpPolicy)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating service edge router policy")
	}

	logrus.Infof("allocated public resources for share '%v' with service id '%v'", shrToken, shrZId)
	return shrZId, frontendEndpoints, nil
}

func (h *share12Handler) allocatePrivateResources(envZId, shrToken string, frontendEndpoints []string, params share.Share12Params, trx interface{}) (string, []string, error) {
	// create automation client
	automationCfg := &automation.Config{
		ApiEndpoint: cfg.Ziti.ApiEndpoint,
		Username:    cfg.Ziti.Username,
		Password:    cfg.Ziti.Password,
	}
	ziti, err := automation.NewZitiAutomation(automationCfg)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating ziti automation client")
	}

	// prepare auth users
	var authUsers []*sdk.AuthUserConfig
	for _, authUser := range params.Body.BasicAuthUsers {
		authUsers = append(authUsers, &sdk.AuthUserConfig{Username: authUser.Username, Password: authUser.Password})
	}

	// parse auth scheme
	authScheme, err := sdk.ParseAuthScheme(params.Body.AuthScheme)
	if err != nil {
		return "", nil, errors.Wrap(err, "error parsing auth scheme")
	}

	// prepare oauth config
	var oauthCfg *sdk.OauthConfig
	if authScheme == sdk.Oauth {
		oauthCfg = &sdk.OauthConfig{
			Provider:                   params.Body.OauthProvider,
			EmailDomains:               params.Body.OauthEmailDomains,
			AuthorizationCheckInterval: params.Body.OauthRefreshInterval,
		}
	}

	// create frontend config (private shares don't use interstitials)
	frontendConfig := &sdk.FrontendConfig{
		Interstitial: false,
		AuthScheme:   authScheme,
	}
	if authScheme == sdk.Basic {
		frontendConfig.BasicAuth = &sdk.BasicAuthConfig{Users: authUsers}
	}
	if authScheme == sdk.Oauth && oauthCfg != nil {
		frontendConfig.OauthAuth = oauthCfg
	}

	// create config using the global zrokProxyConfigId
	tags := automation.ZrokShareTags(shrToken)
	configOpts := &automation.ConfigOptions{
		BaseOptions: automation.BaseOptions{
			Name: shrToken,
			Tags: tags,
		},
		ConfigTypeID: zrokProxyConfigId,
		Data:         frontendConfig,
	}
	cfgZId, err := ziti.Configs.Create(configOpts)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating config")
	}

	// create share service
	serviceOpts := &automation.ServiceOptions{
		BaseOptions: automation.BaseOptions{
			Name: shrToken,
			Tags: tags,
		},
		Configs:            []string{cfgZId},
		EncryptionRequired: true,
	}
	shrZId, err := ziti.Services.Create(serviceOpts)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating share service")
	}

	// create bind policy (backend can bind to this service)
	bindPolicyName := envZId + "-" + shrZId + "-bind"
	bindPolicy := automation.NewPolicyBuilder(bindPolicyName).
		WithServiceIDs(shrZId).
		WithIdentityIDs(envZId).
		WithTags(tags, nil)
	_, err = ziti.Policies.CreateServicePolicyBind(bindPolicy)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating service policy bind")
	}

	// create service edge router policy
	serpPolicyName := envZId + "-" + shrToken + "-serp"
	serpPolicy := automation.NewPolicyBuilder(serpPolicyName).
		WithServiceIDs(shrZId).
		WithAllEdgeRouters().
		WithTags(tags, nil)
	_, err = ziti.Policies.CreateServiceEdgeRouterPolicy(serpPolicy)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating service edge router policy")
	}

	// note: private shares don't create dial policies here
	// dial access is granted separately via the access endpoint

	logrus.Infof("allocated private resources for share '%v' with service id '%v'", shrToken, shrZId)
	return shrZId, frontendEndpoints, nil
}

func (h *share12Handler) createShareRecord(envId int, shrZId, shrToken string, params share.Share12Params, frontendEndpoints []string, trx interface{}) (int, error) {
	strShr := &store.Share{
		ZId:            shrZId,
		Token:          shrToken,
		ShareMode:      params.Body.ShareMode,
		BackendMode:    params.Body.BackendMode,
		Reserved:       false, // share12 doesn't support reserved shares
		UniqueName:     false, // share12 doesn't support unique names
		PermissionMode: store.OpenPermissionMode,
	}

	// set target as backend proxy endpoint for share12
	if params.Body.Target != "" {
		strShr.BackendProxyEndpoint = &params.Body.Target
	}

	// set permission mode if specified
	if params.Body.PermissionMode != "" {
		strShr.PermissionMode = store.PermissionMode(params.Body.PermissionMode)
	}

	// set frontend endpoint (first one if multiple)
	if len(frontendEndpoints) > 0 {
		strShr.FrontendEndpoint = &frontendEndpoints[0]
	} else if strShr.ShareMode == "private" {
		// for private shares without frontend endpoints, use the share mode as endpoint
		strShr.FrontendEndpoint = &strShr.ShareMode
	}

	// create the share record
	shareId, err := str.CreateShare(envId, strShr, trx.(*sqlx.Tx))
	if err != nil {
		return 0, errors.Wrap(err, "error creating share record")
	}

	logrus.Infof("created share record with id '%v' for share '%v'", shareId, shrToken)
	return shareId, nil
}

func (h *share12Handler) processAccessGrants(shareId int, accessGrants []string, permissionMode string, principal *rest_model_zrok.Principal, trx interface{}) error {
	// only process access grants for closed permission mode
	if store.PermissionMode(permissionMode) != store.ClosedPermissionMode {
		return nil
	}

	// find account IDs for the access grant email addresses
	var accessGrantAcctIds []int
	for _, email := range accessGrants {
		acct, err := str.FindAccountWithEmail(email, trx.(*sqlx.Tx))
		if err != nil {
			return errors.Wrapf(err, "unable to find account '%v' for share request from '%v'", email, principal.Email)
		}
		logrus.Debugf("found id '%d' for '%v'", acct.Id, acct.Email)
		accessGrantAcctIds = append(accessGrantAcctIds, acct.Id)
	}

	// create access grants for each account
	for _, acctId := range accessGrantAcctIds {
		_, err := str.CreateAccessGrant(shareId, acctId, trx.(*sqlx.Tx))
		if err != nil {
			return errors.Wrapf(err, "error creating access grant for share '%v' and account '%v'", shareId, acctId)
		}
		logrus.Debugf("created access grant for share '%v' and account '%v'", shareId, acctId)
	}

	if len(accessGrantAcctIds) > 0 {
		logrus.Infof("created %d access grants for closed share '%v'", len(accessGrantAcctIds), shareId)
	}

	return nil
}

func (h *share12Handler) processDynamicMappings(shrToken string, nameIds []int, trx *sqlx.Tx) error {
	// only send updates if dynamic proxy controller is enabled
	if dPCtrl == nil {
		logrus.Warnf("dynamic proxy controller is nil")
		return nil
	}

	for _, nameId := range nameIds {
		logrus.Infof("processing nameId '%v'", nameId)

		// find name record to get the name and namespace
		name, err := str.GetName(nameId, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding name with id '%v'", nameId)
		}
		logrus.Infof("name: %v", name)

		// find namespace
		ns, err := str.GetNamespace(name.NamespaceId, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding namespace with id '%v'", name.NamespaceId)
		}
		logrus.Infof("namespace: %v", ns)

		// find dynamic frontends for this namespace
		frontends, err := str.FindDynamicFrontendsForNamespace(ns.Id, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding dynamic frontends for namespace '%v'", ns.Token)
		}
		logrus.Infof("frontends: %v", frontends)

		// send mapping updates to each dynamic frontend
		for _, frontend := range frontends {
			frontendName := util.ExpandUrlTemplate(name.Name, ns.Name)
			logrus.Infof("binding name '%v'", frontendName)

			if err := dPCtrl.BindFrontendMapping(frontend.Token, frontendName, shrToken, trx); err != nil {
				logrus.Errorf("error binding frontend mapping to frontend '%v': %v", frontend.Token, err)
				// continue with other frontends rather than failing completely
			} else {
				logrus.Infof("bound frontend mapping '%v' to dynamic frontend '%v'", frontendName, frontend.Token)
			}
		}
	}
	return nil
}
