// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openziti/zrok/rest_model_zrok"
)

// InviteCreatedCode is the HTTP code returned for type InviteCreated
const InviteCreatedCode int = 201

/*
InviteCreated invitation created

swagger:response inviteCreated
*/
type InviteCreated struct {
}

// NewInviteCreated creates InviteCreated with default headers values
func NewInviteCreated() *InviteCreated {

	return &InviteCreated{}
}

// WriteResponse to the client
func (o *InviteCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(201)
}

// InviteBadRequestCode is the HTTP code returned for type InviteBadRequest
const InviteBadRequestCode int = 400

/*
InviteBadRequest invitation not created (already exists)

swagger:response inviteBadRequest
*/
type InviteBadRequest struct {

	/*
	  In: Body
	*/
	Payload rest_model_zrok.ErrorMessage `json:"body,omitempty"`
}

// NewInviteBadRequest creates InviteBadRequest with default headers values
func NewInviteBadRequest() *InviteBadRequest {

	return &InviteBadRequest{}
}

// WithPayload adds the payload to the invite bad request response
func (o *InviteBadRequest) WithPayload(payload rest_model_zrok.ErrorMessage) *InviteBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the invite bad request response
func (o *InviteBadRequest) SetPayload(payload rest_model_zrok.ErrorMessage) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *InviteBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// InviteUnauthorizedCode is the HTTP code returned for type InviteUnauthorized
const InviteUnauthorizedCode int = 401

/*
InviteUnauthorized unauthorized

swagger:response inviteUnauthorized
*/
type InviteUnauthorized struct {
}

// NewInviteUnauthorized creates InviteUnauthorized with default headers values
func NewInviteUnauthorized() *InviteUnauthorized {

	return &InviteUnauthorized{}
}

// WriteResponse to the client
func (o *InviteUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// InviteInternalServerErrorCode is the HTTP code returned for type InviteInternalServerError
const InviteInternalServerErrorCode int = 500

/*
InviteInternalServerError internal server error

swagger:response inviteInternalServerError
*/
type InviteInternalServerError struct {
}

// NewInviteInternalServerError creates InviteInternalServerError with default headers values
func NewInviteInternalServerError() *InviteInternalServerError {

	return &InviteInternalServerError{}
}

// WriteResponse to the client
func (o *InviteInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(500)
}
