// Code generated by go-swagger; DO NOT EDIT.

package account

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

// NewVerifyParams creates a new VerifyParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewVerifyParams() *VerifyParams {
	return &VerifyParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewVerifyParamsWithTimeout creates a new VerifyParams object
// with the ability to set a timeout on a request.
func NewVerifyParamsWithTimeout(timeout time.Duration) *VerifyParams {
	return &VerifyParams{
		timeout: timeout,
	}
}

// NewVerifyParamsWithContext creates a new VerifyParams object
// with the ability to set a context for a request.
func NewVerifyParamsWithContext(ctx context.Context) *VerifyParams {
	return &VerifyParams{
		Context: ctx,
	}
}

// NewVerifyParamsWithHTTPClient creates a new VerifyParams object
// with the ability to set a custom HTTPClient for a request.
func NewVerifyParamsWithHTTPClient(client *http.Client) *VerifyParams {
	return &VerifyParams{
		HTTPClient: client,
	}
}

/*
VerifyParams contains all the parameters to send to the API endpoint

	for the verify operation.

	Typically these are written to a http.Request.
*/
type VerifyParams struct {

	// Body.
	Body VerifyBody

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the verify params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *VerifyParams) WithDefaults() *VerifyParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the verify params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *VerifyParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the verify params
func (o *VerifyParams) WithTimeout(timeout time.Duration) *VerifyParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the verify params
func (o *VerifyParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the verify params
func (o *VerifyParams) WithContext(ctx context.Context) *VerifyParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the verify params
func (o *VerifyParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the verify params
func (o *VerifyParams) WithHTTPClient(client *http.Client) *VerifyParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the verify params
func (o *VerifyParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the verify params
func (o *VerifyParams) WithBody(body VerifyBody) *VerifyParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the verify params
func (o *VerifyParams) SetBody(body VerifyBody) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *VerifyParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
