package automation

import (
	"github.com/michaelquigley/df/dl"
	edgeservice "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
)

type ServiceManager struct {
	*BaseResourceManager[rest_model.ServiceDetail]
}

func NewServiceManager(ziti *ZitiAutomation) *ServiceManager {
	return &ServiceManager{
		BaseResourceManager: NewBaseResourceManager[rest_model.ServiceDetail](ziti),
	}
}

// ensure ServiceManager implements the interface
var _ IResourceManager[rest_model.ServiceDetail, *ServiceOptions] = (*ServiceManager)(nil)

type ServiceOptions struct {
	BaseOptions
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

	req := &edgeservice.CreateServiceParams{
		Service: svc,
		Context: sm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := sm.Edge().Service.CreateService(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating service '%s'", opts.Name)
	}

	dl.Infof("created service '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (sm *ServiceManager) Delete(id string) error {
	req := &edgeservice.DeleteServiceParams{
		ID:      id,
		Context: sm.Context(),
	}
	req.SetTimeout(DefaultOperationTimeout)

	_, err := sm.Edge().Service.DeleteService(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting service '%s'", id)
	}

	dl.Infof("deleted service '%s'", id)
	return nil
}

func (sm *ServiceManager) Find(opts *FilterOptions) ([]*rest_model.ServiceDetail, error) {
	req := &edgeservice.ListServicesParams{
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
	return GetByID(sm.Find, id, "service")
}

func (sm *ServiceManager) GetByName(name string) (*rest_model.ServiceDetail, error) {
	return GetByName(sm.Find, name, "service")
}

func (sm *ServiceManager) DeleteWithFilter(filter string) error {
	return DeleteWithFilter(sm.Find, sm.Delete, filter, "service")
}
