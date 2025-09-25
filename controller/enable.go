package controller

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	rest_model_edge "github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/environment"
	"github.com/pkg/errors"
)

type enableHandler struct{}

func newEnableHandler() *enableHandler {
	return &enableHandler{}
}

func (h *enableHandler) Handle(params environment.EnableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	if err := h.checkLimits(principal, trx); err != nil {
		dl.Errorf("limits error for user '%v': %v", principal.Email, err)
		return environment.NewEnableUnauthorized()
	}

	uniqueToken, err := createShareToken()
	if err != nil {
		dl.Errorf("error creating unique identity token for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		dl.Errorf("error getting automation client for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	// create environment identity (equivalent to CreateEnvironmentIdentity)
	identityName := principal.Email + "-" + uniqueToken + "-" + params.Body.Description
	tags := automation.ZrokTags().WithEmail(principal.Email)
	identityOpts := &automation.IdentityOptions{
		BaseOptions: automation.BaseOptions{
			Name: identityName,
			Tags: tags,
		},
		Type:    rest_model_edge.IdentityTypeUser,
		IsAdmin: false,
	}
	envZId, err := ziti.Identities.Create(identityOpts)
	if err != nil {
		dl.Errorf("error creating environment identity for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	// enroll identity
	zitiCfg, err := ziti.Identities.Enroll(envZId)
	if err != nil {
		dl.Errorf("error enrolling environment identity for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	// create edge router policy for the identity
	erpOpts := &automation.EdgeRouterPolicyOptions{
		BaseOptions: automation.BaseOptions{
			Name: envZId,
			Tags: automation.ZrokTags(),
		},
		IdentityRoles:   []string{fmt.Sprintf("@%v", envZId)},
		EdgeRouterRoles: []string{"#all"},
		Semantic:        rest_model_edge.SemanticAllOf,
	}
	if _, err := ziti.EdgeRouterPolicies.Create(erpOpts); err != nil {
		dl.Errorf("error creating edge router policy for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	envId, err := str.CreateEnvironment(int(principal.ID), &store.Environment{
		Description: params.Body.Description,
		Host:        params.Body.Host,
		Address:     realRemoteAddress(params.HTTPRequest),
		ZId:         envZId,
	}, trx)
	if err != nil {
		dl.Errorf("error storing created identity for user '%v': %v", principal.Email, err)
		_ = trx.Rollback()
		return environment.NewEnableInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}
	dl.Infof("created environment for '%v', with ziti identity '%v', and database id '%v'", principal.Email, envZId, envId)

	resp := environment.NewEnableCreated().WithPayload(&environment.EnableCreatedBody{Identity: envZId})

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&zitiCfg)
	if err != nil {
		panic(err)
	}
	resp.Payload.Cfg = out.String()

	return resp
}

func (h *enableHandler) checkLimits(principal *rest_model_zrok.Principal, trx *sqlx.Tx) error {
	if !principal.Limitless {
		if limitsAgent != nil {
			ok, err := limitsAgent.CanCreateEnvironment(int(principal.ID), trx)
			if err != nil {
				return errors.Wrapf(err, "error checking environment limits for '%v'", principal.Email)
			}
			if !ok {
				return errors.Errorf("environment limit check failed for '%v'", principal.Email)
			}
		}
	}
	return nil
}
