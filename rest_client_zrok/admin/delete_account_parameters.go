// Code generated by go-swagger; DO NOT EDIT.

package admin

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

// NewDeleteAccountParams creates a new DeleteAccountParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewDeleteAccountParams() *DeleteAccountParams {
	return &DeleteAccountParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewDeleteAccountParamsWithTimeout creates a new DeleteAccountParams object
// with the ability to set a timeout on a request.
func NewDeleteAccountParamsWithTimeout(timeout time.Duration) *DeleteAccountParams {
	return &DeleteAccountParams{
		timeout: timeout,
	}
}

// NewDeleteAccountParamsWithContext creates a new DeleteAccountParams object
// with the ability to set a context for a request.
func NewDeleteAccountParamsWithContext(ctx context.Context) *DeleteAccountParams {
	return &DeleteAccountParams{
		Context: ctx,
	}
}

// NewDeleteAccountParamsWithHTTPClient creates a new DeleteAccountParams object
// with the ability to set a custom HTTPClient for a request.
func NewDeleteAccountParamsWithHTTPClient(client *http.Client) *DeleteAccountParams {
	return &DeleteAccountParams{
		HTTPClient: client,
	}
}

/*
DeleteAccountParams contains all the parameters to send to the API endpoint

	for the delete account operation.

	Typically these are written to a http.Request.
*/
type DeleteAccountParams struct {

	// Body.
	Body DeleteAccountBody

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the delete account params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteAccountParams) WithDefaults() *DeleteAccountParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the delete account params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteAccountParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the delete account params
func (o *DeleteAccountParams) WithTimeout(timeout time.Duration) *DeleteAccountParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete account params
func (o *DeleteAccountParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete account params
func (o *DeleteAccountParams) WithContext(ctx context.Context) *DeleteAccountParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete account params
func (o *DeleteAccountParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete account params
func (o *DeleteAccountParams) WithHTTPClient(client *http.Client) *DeleteAccountParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete account params
func (o *DeleteAccountParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the delete account params
func (o *DeleteAccountParams) WithBody(body DeleteAccountBody) *DeleteAccountParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the delete account params
func (o *DeleteAccountParams) SetBody(body DeleteAccountBody) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *DeleteAccountParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
