package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
)

type listFrontendNamespaceMappingsHandler struct{}

func newListFrontendNamespaceMappingsHandler() *listFrontendNamespaceMappingsHandler {
	return &listFrontendNamespaceMappingsHandler{}
}

func (handler *listFrontendNamespaceMappingsHandler) Handle(params admin.ListFrontendNamespaceMappingsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewListFrontendNamespaceMappingsUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewListFrontendNamespaceMappingsInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	feToken := params.FrontendToken

	fe, err := str.FindFrontendWithToken(feToken, trx)
	if err != nil {
		dl.Errorf("error finding frontend by token '%s': %v", feToken, err)
		return admin.NewListFrontendNamespaceMappingsNotFound()
	}

	nfMappings, err := str.FindNamespaceFrontendMappingsForFrontend(fe.Id, trx)
	if err != nil {
		dl.Errorf("error finding namespace frontend mappings for frontend '%s': %v", feToken, err)
		return admin.NewListFrontendNamespaceMappingsInternalServerError()
	}

	var mappings []*admin.ListFrontendNamespaceMappingsOKBodyItems0
	for _, nfMapping := range nfMappings {
		ns, err := str.GetNamespace(nfMapping.NamespaceId, trx)
		if err != nil {
			dl.Errorf("error finding namespace with id '%d': %v", nfMapping.NamespaceId, err)
			continue
		}
		mapping := &admin.ListFrontendNamespaceMappingsOKBodyItems0{
			FrontendToken:  feToken,
			NamespaceToken: ns.Token,
			IsDefault:      nfMapping.IsDefault,
			CreatedAt:      nfMapping.CreatedAt.Unix(),
		}
		mappings = append(mappings, mapping)
	}

	dl.Infof("listed '%d' namespace mappings for frontend '%s'", len(mappings), feToken)
	return admin.NewListFrontendNamespaceMappingsOK().WithPayload(mappings)
}
