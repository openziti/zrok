// Code generated by go-swagger; DO NOT EDIT.

package agent

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	"github.com/openziti/zrok/rest_model_zrok"
)

// RemoteShareHandlerFunc turns a function with the right signature into a remote share handler
type RemoteShareHandlerFunc func(RemoteShareParams, *rest_model_zrok.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn RemoteShareHandlerFunc) Handle(params RemoteShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	return fn(params, principal)
}

// RemoteShareHandler interface for that can handle valid remote share params
type RemoteShareHandler interface {
	Handle(RemoteShareParams, *rest_model_zrok.Principal) middleware.Responder
}

// NewRemoteShare creates a new http.Handler for the remote share operation
func NewRemoteShare(ctx *middleware.Context, handler RemoteShareHandler) *RemoteShare {
	return &RemoteShare{Context: ctx, Handler: handler}
}

/*
	RemoteShare swagger:route POST /agent/share agent remoteShare

RemoteShare remote share API
*/
type RemoteShare struct {
	Context *middleware.Context
	Handler RemoteShareHandler
}

func (o *RemoteShare) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewRemoteShareParams()
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

// RemoteShareBody remote share body
//
// swagger:model RemoteShareBody
type RemoteShareBody struct {

	// access grants
	AccessGrants []string `json:"accessGrants"`

	// backend mode
	// Enum: [proxy web tcpTunnel udpTunnel caddy drive socks vpn]
	BackendMode string `json:"backendMode,omitempty"`

	// basic auth
	BasicAuth []string `json:"basicAuth"`

	// env z Id
	EnvZID string `json:"envZId,omitempty"`

	// frontend selection
	FrontendSelection []string `json:"frontendSelection"`

	// insecure
	Insecure bool `json:"insecure,omitempty"`

	// oauth check interval
	OauthCheckInterval string `json:"oauthCheckInterval,omitempty"`

	// oauth email address patterns
	OauthEmailAddressPatterns []string `json:"oauthEmailAddressPatterns"`

	// oauth provider
	OauthProvider string `json:"oauthProvider,omitempty"`

	// open
	Open bool `json:"open,omitempty"`

	// share mode
	// Enum: [public private reserved]
	ShareMode string `json:"shareMode,omitempty"`

	// target
	Target string `json:"target,omitempty"`

	// token
	Token string `json:"token,omitempty"`
}

// Validate validates this remote share body
func (o *RemoteShareBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateBackendMode(formats); err != nil {
		res = append(res, err)
	}

	if err := o.validateShareMode(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var remoteShareBodyTypeBackendModePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["proxy","web","tcpTunnel","udpTunnel","caddy","drive","socks","vpn"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		remoteShareBodyTypeBackendModePropEnum = append(remoteShareBodyTypeBackendModePropEnum, v)
	}
}

const (

	// RemoteShareBodyBackendModeProxy captures enum value "proxy"
	RemoteShareBodyBackendModeProxy string = "proxy"

	// RemoteShareBodyBackendModeWeb captures enum value "web"
	RemoteShareBodyBackendModeWeb string = "web"

	// RemoteShareBodyBackendModeTCPTunnel captures enum value "tcpTunnel"
	RemoteShareBodyBackendModeTCPTunnel string = "tcpTunnel"

	// RemoteShareBodyBackendModeUDPTunnel captures enum value "udpTunnel"
	RemoteShareBodyBackendModeUDPTunnel string = "udpTunnel"

	// RemoteShareBodyBackendModeCaddy captures enum value "caddy"
	RemoteShareBodyBackendModeCaddy string = "caddy"

	// RemoteShareBodyBackendModeDrive captures enum value "drive"
	RemoteShareBodyBackendModeDrive string = "drive"

	// RemoteShareBodyBackendModeSocks captures enum value "socks"
	RemoteShareBodyBackendModeSocks string = "socks"

	// RemoteShareBodyBackendModeVpn captures enum value "vpn"
	RemoteShareBodyBackendModeVpn string = "vpn"
)

// prop value enum
func (o *RemoteShareBody) validateBackendModeEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, remoteShareBodyTypeBackendModePropEnum, true); err != nil {
		return err
	}
	return nil
}

func (o *RemoteShareBody) validateBackendMode(formats strfmt.Registry) error {
	if swag.IsZero(o.BackendMode) { // not required
		return nil
	}

	// value enum
	if err := o.validateBackendModeEnum("body"+"."+"backendMode", "body", o.BackendMode); err != nil {
		return err
	}

	return nil
}

var remoteShareBodyTypeShareModePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["public","private","reserved"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		remoteShareBodyTypeShareModePropEnum = append(remoteShareBodyTypeShareModePropEnum, v)
	}
}

const (

	// RemoteShareBodyShareModePublic captures enum value "public"
	RemoteShareBodyShareModePublic string = "public"

	// RemoteShareBodyShareModePrivate captures enum value "private"
	RemoteShareBodyShareModePrivate string = "private"

	// RemoteShareBodyShareModeReserved captures enum value "reserved"
	RemoteShareBodyShareModeReserved string = "reserved"
)

// prop value enum
func (o *RemoteShareBody) validateShareModeEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, remoteShareBodyTypeShareModePropEnum, true); err != nil {
		return err
	}
	return nil
}

func (o *RemoteShareBody) validateShareMode(formats strfmt.Registry) error {
	if swag.IsZero(o.ShareMode) { // not required
		return nil
	}

	// value enum
	if err := o.validateShareModeEnum("body"+"."+"shareMode", "body", o.ShareMode); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this remote share body based on context it is used
func (o *RemoteShareBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RemoteShareBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RemoteShareBody) UnmarshalBinary(b []byte) error {
	var res RemoteShareBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

// RemoteShareOKBody remote share o k body
//
// swagger:model RemoteShareOKBody
type RemoteShareOKBody struct {

	// frontend endpoints
	FrontendEndpoints []string `json:"frontendEndpoints"`

	// token
	Token string `json:"token,omitempty"`
}

// Validate validates this remote share o k body
func (o *RemoteShareOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this remote share o k body based on context it is used
func (o *RemoteShareOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RemoteShareOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RemoteShareOKBody) UnmarshalBinary(b []byte) error {
	var res RemoteShareOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
