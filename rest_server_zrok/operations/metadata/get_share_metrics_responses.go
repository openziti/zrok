// Code generated by go-swagger; DO NOT EDIT.

package metadata

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openziti/zrok/rest_model_zrok"
)

// GetShareMetricsOKCode is the HTTP code returned for type GetShareMetricsOK
const GetShareMetricsOKCode int = 200

/*
GetShareMetricsOK share metrics

swagger:response getShareMetricsOK
*/
type GetShareMetricsOK struct {

	/*
	  In: Body
	*/
	Payload *rest_model_zrok.Metrics `json:"body,omitempty"`
}

// NewGetShareMetricsOK creates GetShareMetricsOK with default headers values
func NewGetShareMetricsOK() *GetShareMetricsOK {

	return &GetShareMetricsOK{}
}

// WithPayload adds the payload to the get share metrics o k response
func (o *GetShareMetricsOK) WithPayload(payload *rest_model_zrok.Metrics) *GetShareMetricsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get share metrics o k response
func (o *GetShareMetricsOK) SetPayload(payload *rest_model_zrok.Metrics) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetShareMetricsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetShareMetricsBadRequestCode is the HTTP code returned for type GetShareMetricsBadRequest
const GetShareMetricsBadRequestCode int = 400

/*
GetShareMetricsBadRequest bad request

swagger:response getShareMetricsBadRequest
*/
type GetShareMetricsBadRequest struct {
}

// NewGetShareMetricsBadRequest creates GetShareMetricsBadRequest with default headers values
func NewGetShareMetricsBadRequest() *GetShareMetricsBadRequest {

	return &GetShareMetricsBadRequest{}
}

// WriteResponse to the client
func (o *GetShareMetricsBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// GetShareMetricsUnauthorizedCode is the HTTP code returned for type GetShareMetricsUnauthorized
const GetShareMetricsUnauthorizedCode int = 401

/*
GetShareMetricsUnauthorized unauthorized

swagger:response getShareMetricsUnauthorized
*/
type GetShareMetricsUnauthorized struct {
}

// NewGetShareMetricsUnauthorized creates GetShareMetricsUnauthorized with default headers values
func NewGetShareMetricsUnauthorized() *GetShareMetricsUnauthorized {

	return &GetShareMetricsUnauthorized{}
}

// WriteResponse to the client
func (o *GetShareMetricsUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// GetShareMetricsInternalServerErrorCode is the HTTP code returned for type GetShareMetricsInternalServerError
const GetShareMetricsInternalServerErrorCode int = 500

/*
GetShareMetricsInternalServerError internal server error

swagger:response getShareMetricsInternalServerError
*/
type GetShareMetricsInternalServerError struct {
}

// NewGetShareMetricsInternalServerError creates GetShareMetricsInternalServerError with default headers values
func NewGetShareMetricsInternalServerError() *GetShareMetricsInternalServerError {

	return &GetShareMetricsInternalServerError{}
}

// WriteResponse to the client
func (o *GetShareMetricsInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(500)
}
