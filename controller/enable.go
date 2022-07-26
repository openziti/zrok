package controller

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations/identity"
	"github.com/openziti/edge/rest_management_api_client"
	identity_edge "github.com/openziti/edge/rest_management_api_client/identity"
	rest_model_edge "github.com/openziti/edge/rest_model"
	"github.com/openziti/edge/rest_util"
	sdk_config "github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func enableHandler(params identity.EnableParams) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return middleware.Error(500, err.Error())
	}
	a, err := str.FindAccountWithToken(params.Body.Token, tx)
	if err != nil {
		logrus.Errorf("error finding account: %v", err)
		return middleware.Error(500, err.Error())
	}
	if a == nil {
		logrus.Errorf("account not found: %v", err)
		return middleware.Error(404, err.Error())
	}
	logrus.Infof("found account '%v'", a.Username)

	ctrlAddress := "https://linux:1280"
	caCerts, err := rest_util.GetControllerWellKnownCas(ctrlAddress)
	if err != nil {
		panic(errors.Wrap(err, "error getting cas"))
	}
	caPool := x509.NewCertPool()
	for _, ca := range caCerts {
		caPool.AddCert(ca)
	}
	client, err := rest_util.NewEdgeManagementClientWithUpdb("admin", "admin", ctrlAddress, caPool)
	if err != nil {
		panic(err)
	}
	ident, err := createIdentity(a, client)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	cfg, err := enrollIdentity(ident.Payload.Data.ID, client)
	if err != nil {
		panic(err)
	}

	resp := identity.NewEnableCreated().WithPayload(&rest_model.EnableResponse{
		Identity: ident.Payload.Data.ID,
	})

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&cfg)
	if err != nil {
		panic(err)
	}
	resp.Payload.Cfg = out.String()

	return resp
}

func createIdentity(a *store.Account, client *rest_management_api_client.ZitiEdgeManagement) (*identity_edge.CreateIdentityCreated, error) {
	iIsAdmin := false
	iId, err := generateIdentityId()
	if err != nil {
		return nil, err
	}
	iName := fmt.Sprintf("%v-%v", a.Username, iId)
	iType := rest_model_edge.IdentityTypeUser
	i := &rest_model_edge.IdentityCreate{
		Enrollment:          &rest_model_edge.IdentityCreateEnrollment{Ott: true},
		IsAdmin:             &iIsAdmin,
		Name:                &iName,
		RoleAttributes:      nil,
		ServiceHostingCosts: nil,
		Tags:                nil,
		Type:                &iType,
	}
	p := identity_edge.NewCreateIdentityParams()
	p.Identity = i
	ident, err := client.Identity.CreateIdentity(p, nil)
	if err != nil {
		return nil, err
	}
	return ident, nil
}

func enrollIdentity(id string, client *rest_management_api_client.ZitiEdgeManagement) (*sdk_config.Config, error) {
	p := &identity_edge.DetailIdentityParams{
		Context: context.Background(),
		ID:      id,
	}
	p.SetTimeout(30 * time.Second)
	resp, err := client.Identity.DetailIdentity(p, nil)
	if err != nil {
		return nil, err
	}
	tkn, _, err := enroll.ParseToken(resp.GetPayload().Data.Enrollment.Ott.JWT)
	if err != nil {
		return nil, err
	}
	flags := enroll.EnrollmentFlags{
		Token:  tkn,
		KeyAlg: "RSA",
	}
	conf, err := enroll.Enroll(flags)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
