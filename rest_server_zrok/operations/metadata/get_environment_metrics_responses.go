// Code generated by go-swagger; DO NOT EDIT.

package metadata

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openziti/zrok/rest_model_zrok"
)

// GetEnvironmentMetricsOKCode is the HTTP code returned for type GetEnvironmentMetricsOK
const GetEnvironmentMetricsOKCode int = 200

/*
GetEnvironmentMetricsOK environment metrics

swagger:response getEnvironmentMetricsOK
*/
type GetEnvironmentMetricsOK struct {

	/*
	  In: Body
	*/
	Payload *rest_model_zrok.Metrics `json:"body,omitempty"`
}

// NewGetEnvironmentMetricsOK creates GetEnvironmentMetricsOK with default headers values
func NewGetEnvironmentMetricsOK() *GetEnvironmentMetricsOK {

	return &GetEnvironmentMetricsOK{}
}

// WithPayload adds the payload to the get environment metrics o k response
func (o *GetEnvironmentMetricsOK) WithPayload(payload *rest_model_zrok.Metrics) *GetEnvironmentMetricsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get environment metrics o k response
func (o *GetEnvironmentMetricsOK) SetPayload(payload *rest_model_zrok.Metrics) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetEnvironmentMetricsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetEnvironmentMetricsBadRequestCode is the HTTP code returned for type GetEnvironmentMetricsBadRequest
const GetEnvironmentMetricsBadRequestCode int = 400

/*
GetEnvironmentMetricsBadRequest bad request

swagger:response getEnvironmentMetricsBadRequest
*/
type GetEnvironmentMetricsBadRequest struct {
}

// NewGetEnvironmentMetricsBadRequest creates GetEnvironmentMetricsBadRequest with default headers values
func NewGetEnvironmentMetricsBadRequest() *GetEnvironmentMetricsBadRequest {

	return &GetEnvironmentMetricsBadRequest{}
}

// WriteResponse to the client
func (o *GetEnvironmentMetricsBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// GetEnvironmentMetricsUnauthorizedCode is the HTTP code returned for type GetEnvironmentMetricsUnauthorized
const GetEnvironmentMetricsUnauthorizedCode int = 401

/*
GetEnvironmentMetricsUnauthorized unauthorized

swagger:response getEnvironmentMetricsUnauthorized
*/
type GetEnvironmentMetricsUnauthorized struct {
}

// NewGetEnvironmentMetricsUnauthorized creates GetEnvironmentMetricsUnauthorized with default headers values
func NewGetEnvironmentMetricsUnauthorized() *GetEnvironmentMetricsUnauthorized {

	return &GetEnvironmentMetricsUnauthorized{}
}

// WriteResponse to the client
func (o *GetEnvironmentMetricsUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// GetEnvironmentMetricsInternalServerErrorCode is the HTTP code returned for type GetEnvironmentMetricsInternalServerError
const GetEnvironmentMetricsInternalServerErrorCode int = 500

/*
GetEnvironmentMetricsInternalServerError internal server error

swagger:response getEnvironmentMetricsInternalServerError
*/
type GetEnvironmentMetricsInternalServerError struct {
}

// NewGetEnvironmentMetricsInternalServerError creates GetEnvironmentMetricsInternalServerError with default headers values
func NewGetEnvironmentMetricsInternalServerError() *GetEnvironmentMetricsInternalServerError {

	return &GetEnvironmentMetricsInternalServerError{}
}

// WriteResponse to the client
func (o *GetEnvironmentMetricsInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(500)
}
