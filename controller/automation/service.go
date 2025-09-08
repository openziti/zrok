package automation

import (
	"time"

	edge_service "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ServiceManager struct {
	*ResourceManager
}

func NewServiceManager(client *Client) *ServiceManager {
	return &ServiceManager{
		ResourceManager: NewResourceManager(client),
	}
}

type ServiceOptions struct {
	*ResourceOptions
	Configs            []string
	EncryptionRequired bool
	TerminatorStrategy string
	RoleAttributes     []string
	MaxIdleTime        *int64
}

func (sm *ServiceManager) Create(opts *ServiceOptions) (string, error) {
	svc := &rest_model.ServiceCreate{
		EncryptionRequired: &opts.EncryptionRequired,
		Name:               &opts.Name,
		Tags:               opts.GetTags(),
	}

	if opts.Configs != nil {
		svc.Configs = opts.Configs
	}

	if opts.TerminatorStrategy != "" {
		svc.TerminatorStrategy = opts.TerminatorStrategy
	}

	if opts.RoleAttributes != nil {
		svc.RoleAttributes = opts.RoleAttributes
	}

	if opts.MaxIdleTime != nil {
		svc.MaxIdleTimeMillis = *opts.MaxIdleTime
	}

	req := &edge_service.CreateServiceParams{
		Service: svc,
		Context: sm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := sm.Edge().Service.CreateService(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating service '%s'", opts.Name)
	}

	logrus.Infof("created service '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (sm *ServiceManager) Delete(id string) error {
	req := &edge_service.DeleteServiceParams{
		ID:      id,
		Context: sm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	_, err := sm.Edge().Service.DeleteService(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting service '%s'", id)
	}

	logrus.Infof("deleted service '%s'", id)
	return nil
}

func (sm *ServiceManager) Find(opts *FilterOptions) ([]*rest_model.ServiceDetail, error) {
	req := &edge_service.ListServicesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: sm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := sm.Edge().Service.ListServices(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing services")
	}

	return resp.Payload.Data, nil
}

func (sm *ServiceManager) GetByID(id string) (*rest_model.ServiceDetail, error) {
	opts := &FilterOptions{Filter: BuildFilter("id", id)}
	services, err := sm.Find(opts)
	if err != nil {
		return nil, err
	}
	if len(services) != 1 {
		return nil, errors.Errorf("expected 1 service, found %d", len(services))
	}
	return services[0], nil
}

func (sm *ServiceManager) GetByName(name string) (*rest_model.ServiceDetail, error) {
	opts := &FilterOptions{Filter: BuildFilter("name", name)}
	services, err := sm.Find(opts)
	if err != nil {
		return nil, err
	}
	if len(services) != 1 {
		return nil, errors.Errorf("expected 1 service with name '%s', found %d", name, len(services))
	}
	return services[0], nil
}
