// Code generated by go-swagger; DO NOT EDIT.

package share

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/openziti/zrok/rest_model_zrok"
)

// UpdateShareHandlerFunc turns a function with the right signature into a update share handler
type UpdateShareHandlerFunc func(UpdateShareParams, *rest_model_zrok.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn UpdateShareHandlerFunc) Handle(params UpdateShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	return fn(params, principal)
}

// UpdateShareHandler interface for that can handle valid update share params
type UpdateShareHandler interface {
	Handle(UpdateShareParams, *rest_model_zrok.Principal) middleware.Responder
}

// NewUpdateShare creates a new http.Handler for the update share operation
func NewUpdateShare(ctx *middleware.Context, handler UpdateShareHandler) *UpdateShare {
	return &UpdateShare{Context: ctx, Handler: handler}
}

/*
	UpdateShare swagger:route PATCH /share share updateShare

UpdateShare update share API
*/
type UpdateShare struct {
	Context *middleware.Context
	Handler UpdateShareHandler
}

func (o *UpdateShare) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewUpdateShareParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal *rest_model_zrok.Principal
	if uprinc != nil {
		principal = uprinc.(*rest_model_zrok.Principal) // this is really a rest_model_zrok.Principal, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
