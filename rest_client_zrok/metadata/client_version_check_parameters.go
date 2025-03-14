// Code generated by go-swagger; DO NOT EDIT.

package metadata

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewClientVersionCheckParams creates a new ClientVersionCheckParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewClientVersionCheckParams() *ClientVersionCheckParams {
	return &ClientVersionCheckParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewClientVersionCheckParamsWithTimeout creates a new ClientVersionCheckParams object
// with the ability to set a timeout on a request.
func NewClientVersionCheckParamsWithTimeout(timeout time.Duration) *ClientVersionCheckParams {
	return &ClientVersionCheckParams{
		timeout: timeout,
	}
}

// NewClientVersionCheckParamsWithContext creates a new ClientVersionCheckParams object
// with the ability to set a context for a request.
func NewClientVersionCheckParamsWithContext(ctx context.Context) *ClientVersionCheckParams {
	return &ClientVersionCheckParams{
		Context: ctx,
	}
}

// NewClientVersionCheckParamsWithHTTPClient creates a new ClientVersionCheckParams object
// with the ability to set a custom HTTPClient for a request.
func NewClientVersionCheckParamsWithHTTPClient(client *http.Client) *ClientVersionCheckParams {
	return &ClientVersionCheckParams{
		HTTPClient: client,
	}
}

/*
ClientVersionCheckParams contains all the parameters to send to the API endpoint

	for the client version check operation.

	Typically these are written to a http.Request.
*/
type ClientVersionCheckParams struct {

	// Body.
	Body ClientVersionCheckBody

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the client version check params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ClientVersionCheckParams) WithDefaults() *ClientVersionCheckParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the client version check params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ClientVersionCheckParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the client version check params
func (o *ClientVersionCheckParams) WithTimeout(timeout time.Duration) *ClientVersionCheckParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the client version check params
func (o *ClientVersionCheckParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the client version check params
func (o *ClientVersionCheckParams) WithContext(ctx context.Context) *ClientVersionCheckParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the client version check params
func (o *ClientVersionCheckParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the client version check params
func (o *ClientVersionCheckParams) WithHTTPClient(client *http.Client) *ClientVersionCheckParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the client version check params
func (o *ClientVersionCheckParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the client version check params
func (o *ClientVersionCheckParams) WithBody(body ClientVersionCheckBody) *ClientVersionCheckParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the client version check params
func (o *ClientVersionCheckParams) SetBody(body ClientVersionCheckBody) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *ClientVersionCheckParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
