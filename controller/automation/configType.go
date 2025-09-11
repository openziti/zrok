package automation

import (
	"time"

	"github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConfigTypeManager struct {
	*BaseResourceManager[rest_model.ConfigTypeDetail]
}

func NewConfigTypeManager(client *Client) *ConfigTypeManager {
	return &ConfigTypeManager{
		BaseResourceManager: NewBaseResourceManager[rest_model.ConfigTypeDetail](client),
	}
}

type ConfigTypeOptions struct {
	BaseOptions
	Schema interface{}
}

func (ctm *ConfigTypeManager) Create(opts *ConfigTypeOptions) (string, error) {
	ct := &rest_model.ConfigTypeCreate{
		Name:   &opts.Name,
		Schema: opts.Schema,
		Tags:   opts.GetTags(),
	}

	req := &config.CreateConfigTypeParams{
		ConfigType: ct,
		Context:    ctm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := ctm.Edge().Config.CreateConfigType(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating config type '%s'", opts.Name)
	}

	logrus.Infof("created config type '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (ctm *ConfigTypeManager) Delete(id string) error {
	req := &config.DeleteConfigTypeParams{
		ID:      id,
		Context: ctm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	_, err := ctm.Edge().Config.DeleteConfigType(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting config type '%s'", id)
	}

	logrus.Infof("deleted config type '%s'", id)
	return nil
}

func (ctm *ConfigTypeManager) Find(opts *FilterOptions) ([]*rest_model.ConfigTypeDetail, error) {
	req := &config.ListConfigTypesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: ctm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := ctm.Edge().Config.ListConfigTypes(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing config types")
	}

	return resp.Payload.Data, nil
}

func (ctm *ConfigTypeManager) GetByID(id string) (*rest_model.ConfigTypeDetail, error) {
	return GetByID(ctm.Find, id, "config type")
}

func (ctm *ConfigTypeManager) GetByName(name string) (*rest_model.ConfigTypeDetail, error) {
	return GetByName(ctm.Find, name, "config type")
}

func (ctm *ConfigTypeManager) DeleteWithFilter(filter string) error {
	return DeleteWithFilter(ctm.Find, ctm.Delete, filter, "config type")
}

func (ctm *ConfigTypeManager) EnsureExists(name string) (string, error) {
	existing, err := ctm.GetByName(name)
	if err != nil {
		return "", err
	}

	if existing != nil {
		logrus.Infof("found existing config type '%s' with id '%s'", name, *existing.ID)
		return *existing.ID, nil
	}

	// create it
	opts := &ConfigTypeOptions{
		BaseOptions: BaseOptions{
			Name: name,
		},
		Schema: nil, // no schema for zrok proxy config
	}

	id, err := ctm.Create(opts)
	if err != nil {
		return "", err
	}

	return id, nil
}

// ensure ConfigTypeManager implements the interface
var _ IResourceManager[rest_model.ConfigTypeDetail, *ConfigTypeOptions] = (*ConfigTypeManager)(nil)
