package automation

import (
	"time"

	"github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConfigManager struct {
	*ResourceManager
}

func NewConfigManager(client *Client) *ConfigManager {
	return &ConfigManager{
		ResourceManager: NewResourceManager(client),
	}
}

type ConfigOptions struct {
	*ResourceOptions
	ConfigTypeID string
	Data         interface{}
}

func (cm *ConfigManager) Create(opts *ConfigOptions) (string, error) {
	cfg := &rest_model.ConfigCreate{
		ConfigTypeID: &opts.ConfigTypeID,
		Data:         opts.Data,
		Name:         &opts.Name,
		Tags:         opts.GetTags(),
	}

	req := &config.CreateConfigParams{
		Config:  cfg,
		Context: cm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := cm.Edge().Config.CreateConfig(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating config '%s'", opts.Name)
	}

	logrus.Infof("created config '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (cm *ConfigManager) Update(id string, opts *ConfigOptions) error {
	req := &config.UpdateConfigParams{
		Config: &rest_model.ConfigUpdate{
			Data: opts.Data,
			Name: &opts.Name,
			Tags: opts.GetTags(),
		},
		ID:      id,
		Context: cm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	_, err := cm.Edge().Config.UpdateConfig(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error updating config '%s'", id)
	}

	logrus.Infof("updated config '%s'", id)
	return nil
}

func (cm *ConfigManager) Delete(id string) error {
	req := &config.DeleteConfigParams{
		ID:      id,
		Context: cm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	_, err := cm.Edge().Config.DeleteConfig(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting config '%s'", id)
	}

	logrus.Infof("deleted config '%s'", id)
	return nil
}

func (cm *ConfigManager) Find(opts *FilterOptions) ([]*rest_model.ConfigDetail, error) {
	req := &config.ListConfigsParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: cm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := cm.Edge().Config.ListConfigs(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing configs")
	}

	return resp.Payload.Data, nil
}

func (cm *ConfigManager) GetByID(id string) (*rest_model.ConfigDetail, error) {
	opts := &FilterOptions{Filter: BuildFilter("id", id)}
	configs, err := cm.Find(opts)
	if err != nil {
		return nil, err
	}
	if len(configs) != 1 {
		return nil, errors.Errorf("expected 1 config, found %d", len(configs))
	}
	return configs[0], nil
}

func (cm *ConfigManager) DeleteWithFilter(filter string) error {
	opts := &FilterOptions{Filter: filter}
	configs, err := cm.Find(opts)
	if err != nil {
		return err
	}

	logrus.Infof("found %d configs to delete for filter '%s'", len(configs), filter)

	for _, cfg := range configs {
		if err := cm.Delete(*cfg.ID); err != nil {
			return err
		}
	}

	if len(configs) == 0 {
		logrus.Warnf("no configs found for filter '%s'", filter)
	}

	return nil
}
