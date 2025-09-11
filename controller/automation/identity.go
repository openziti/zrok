package automation

import (
	"time"

	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type IdentityManager struct {
	*BaseResourceManager[rest_model.IdentityDetail]
}

func NewIdentityManager(client *Client) *IdentityManager {
	return &IdentityManager{
		BaseResourceManager: NewBaseResourceManager[rest_model.IdentityDetail](client),
	}
}

type IdentityOptions struct {
	BaseOptions
	Type           rest_model.IdentityType
	IsAdmin        bool
	RoleAttributes []string
}

func (im *IdentityManager) Create(opts *IdentityOptions) (string, error) {
	req := identity.NewCreateIdentityParams()
	req.Identity = &rest_model.IdentityCreate{
		Enrollment:          &rest_model.IdentityCreateEnrollment{Ott: true},
		IsAdmin:             &opts.IsAdmin,
		Name:                &opts.Name,
		RoleAttributes:      (*rest_model.Attributes)(&opts.RoleAttributes),
		ServiceHostingCosts: nil,
		Tags:                opts.GetTags(),
		Type:                &opts.Type,
	}
	req.SetTimeout(opts.GetTimeout())
	req.Context = im.Context()

	resp, err := im.Edge().Identity.CreateIdentity(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating identity '%s'", opts.Name)
	}

	logrus.Infof("created identity '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (im *IdentityManager) Delete(id string) error {
	req := &identity.DeleteIdentityParams{
		ID:      id,
		Context: im.Context(),
	}
	req.SetTimeout(30 * time.Second)

	_, err := im.Edge().Identity.DeleteIdentity(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting identity '%s'", id)
	}

	logrus.Infof("deleted identity '%s'", id)
	return nil
}

func (im *IdentityManager) Find(opts *FilterOptions) ([]*rest_model.IdentityDetail, error) {
	req := &identity.ListIdentitiesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: im.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := im.Edge().Identity.ListIdentities(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing identities")
	}

	return resp.Payload.Data, nil
}

func (im *IdentityManager) GetByID(id string) (*rest_model.IdentityDetail, error) {
	return GetByID(im.Find, id, "identity")
}

func (im *IdentityManager) GetByName(name string) (*rest_model.IdentityDetail, error) {
	return GetByName(im.Find, name, "identity")
}

func (im *IdentityManager) DeleteWithFilter(filter string) error {
	return DeleteWithFilter(im.Find, im.Delete, filter, "identity")
}

func (im *IdentityManager) Enroll(id string) (*ziti.Config, error) {
	p := &identity.DetailIdentityParams{
		Context: im.Context(),
		ID:      id,
	}
	p.SetTimeout(30 * time.Second)

	resp, err := im.Edge().Identity.DetailIdentity(p, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting identity details for '%s'", id)
	}

	tkn, _, err := enroll.ParseToken(resp.GetPayload().Data.Enrollment.Ott.JWT)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing enrollment token")
	}

	flags := enroll.EnrollmentFlags{
		Token:  tkn,
		KeyAlg: "RSA",
	}

	conf, err := enroll.Enroll(flags)
	if err != nil {
		return nil, errors.Wrap(err, "error enrolling identity")
	}

	logrus.Infof("enrolled identity '%s'", id)
	return conf, nil
}

// ensure IdentityManager implements the interface
var _ IResourceManager[rest_model.IdentityDetail, *IdentityOptions] = (*IdentityManager)(nil)
