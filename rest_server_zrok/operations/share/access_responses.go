// Code generated by go-swagger; DO NOT EDIT.

package share

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// AccessCreatedCode is the HTTP code returned for type AccessCreated
const AccessCreatedCode int = 201

/*
AccessCreated access created

swagger:response accessCreated
*/
type AccessCreated struct {

	/*
	  In: Body
	*/
	Payload *AccessCreatedBody `json:"body,omitempty"`
}

// NewAccessCreated creates AccessCreated with default headers values
func NewAccessCreated() *AccessCreated {

	return &AccessCreated{}
}

// WithPayload adds the payload to the access created response
func (o *AccessCreated) WithPayload(payload *AccessCreatedBody) *AccessCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the access created response
func (o *AccessCreated) SetPayload(payload *AccessCreatedBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *AccessCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// AccessUnauthorizedCode is the HTTP code returned for type AccessUnauthorized
const AccessUnauthorizedCode int = 401

/*
AccessUnauthorized unauthorized

swagger:response accessUnauthorized
*/
type AccessUnauthorized struct {
}

// NewAccessUnauthorized creates AccessUnauthorized with default headers values
func NewAccessUnauthorized() *AccessUnauthorized {

	return &AccessUnauthorized{}
}

// WriteResponse to the client
func (o *AccessUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// AccessNotFoundCode is the HTTP code returned for type AccessNotFound
const AccessNotFoundCode int = 404

/*
AccessNotFound not found

swagger:response accessNotFound
*/
type AccessNotFound struct {
}

// NewAccessNotFound creates AccessNotFound with default headers values
func NewAccessNotFound() *AccessNotFound {

	return &AccessNotFound{}
}

// WriteResponse to the client
func (o *AccessNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}

// AccessInternalServerErrorCode is the HTTP code returned for type AccessInternalServerError
const AccessInternalServerErrorCode int = 500

/*
AccessInternalServerError internal server error

swagger:response accessInternalServerError
*/
type AccessInternalServerError struct {
}

// NewAccessInternalServerError creates AccessInternalServerError with default headers values
func NewAccessInternalServerError() *AccessInternalServerError {

	return &AccessInternalServerError{}
}

// WriteResponse to the client
func (o *AccessInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(500)
}
