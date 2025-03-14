// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// CreateAccountReader is a Reader for the CreateAccount structure.
type CreateAccountReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateAccountReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewCreateAccountCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewCreateAccountUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewCreateAccountInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[POST /account] createAccount", response, response.Code())
	}
}

// NewCreateAccountCreated creates a CreateAccountCreated with default headers values
func NewCreateAccountCreated() *CreateAccountCreated {
	return &CreateAccountCreated{}
}

/*
CreateAccountCreated describes a response with status code 201, with default header values.

created
*/
type CreateAccountCreated struct {
	Payload *CreateAccountCreatedBody
}

// IsSuccess returns true when this create account created response has a 2xx status code
func (o *CreateAccountCreated) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this create account created response has a 3xx status code
func (o *CreateAccountCreated) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create account created response has a 4xx status code
func (o *CreateAccountCreated) IsClientError() bool {
	return false
}

// IsServerError returns true when this create account created response has a 5xx status code
func (o *CreateAccountCreated) IsServerError() bool {
	return false
}

// IsCode returns true when this create account created response a status code equal to that given
func (o *CreateAccountCreated) IsCode(code int) bool {
	return code == 201
}

// Code gets the status code for the create account created response
func (o *CreateAccountCreated) Code() int {
	return 201
}

func (o *CreateAccountCreated) Error() string {
	return fmt.Sprintf("[POST /account][%d] createAccountCreated  %+v", 201, o.Payload)
}

func (o *CreateAccountCreated) String() string {
	return fmt.Sprintf("[POST /account][%d] createAccountCreated  %+v", 201, o.Payload)
}

func (o *CreateAccountCreated) GetPayload() *CreateAccountCreatedBody {
	return o.Payload
}

func (o *CreateAccountCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(CreateAccountCreatedBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateAccountUnauthorized creates a CreateAccountUnauthorized with default headers values
func NewCreateAccountUnauthorized() *CreateAccountUnauthorized {
	return &CreateAccountUnauthorized{}
}

/*
CreateAccountUnauthorized describes a response with status code 401, with default header values.

unauthorized
*/
type CreateAccountUnauthorized struct {
}

// IsSuccess returns true when this create account unauthorized response has a 2xx status code
func (o *CreateAccountUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this create account unauthorized response has a 3xx status code
func (o *CreateAccountUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create account unauthorized response has a 4xx status code
func (o *CreateAccountUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this create account unauthorized response has a 5xx status code
func (o *CreateAccountUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this create account unauthorized response a status code equal to that given
func (o *CreateAccountUnauthorized) IsCode(code int) bool {
	return code == 401
}

// Code gets the status code for the create account unauthorized response
func (o *CreateAccountUnauthorized) Code() int {
	return 401
}

func (o *CreateAccountUnauthorized) Error() string {
	return fmt.Sprintf("[POST /account][%d] createAccountUnauthorized ", 401)
}

func (o *CreateAccountUnauthorized) String() string {
	return fmt.Sprintf("[POST /account][%d] createAccountUnauthorized ", 401)
}

func (o *CreateAccountUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewCreateAccountInternalServerError creates a CreateAccountInternalServerError with default headers values
func NewCreateAccountInternalServerError() *CreateAccountInternalServerError {
	return &CreateAccountInternalServerError{}
}

/*
CreateAccountInternalServerError describes a response with status code 500, with default header values.

internal server error
*/
type CreateAccountInternalServerError struct {
}

// IsSuccess returns true when this create account internal server error response has a 2xx status code
func (o *CreateAccountInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this create account internal server error response has a 3xx status code
func (o *CreateAccountInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create account internal server error response has a 4xx status code
func (o *CreateAccountInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this create account internal server error response has a 5xx status code
func (o *CreateAccountInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this create account internal server error response a status code equal to that given
func (o *CreateAccountInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the create account internal server error response
func (o *CreateAccountInternalServerError) Code() int {
	return 500
}

func (o *CreateAccountInternalServerError) Error() string {
	return fmt.Sprintf("[POST /account][%d] createAccountInternalServerError ", 500)
}

func (o *CreateAccountInternalServerError) String() string {
	return fmt.Sprintf("[POST /account][%d] createAccountInternalServerError ", 500)
}

func (o *CreateAccountInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

/*
CreateAccountBody create account body
swagger:model CreateAccountBody
*/
type CreateAccountBody struct {

	// email
	Email string `json:"email,omitempty"`

	// password
	Password string `json:"password,omitempty"`
}

// Validate validates this create account body
func (o *CreateAccountBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this create account body based on context it is used
func (o *CreateAccountBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *CreateAccountBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *CreateAccountBody) UnmarshalBinary(b []byte) error {
	var res CreateAccountBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*
CreateAccountCreatedBody create account created body
swagger:model CreateAccountCreatedBody
*/
type CreateAccountCreatedBody struct {

	// account token
	AccountToken string `json:"accountToken,omitempty"`
}

// Validate validates this create account created body
func (o *CreateAccountCreatedBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this create account created body based on context it is used
func (o *CreateAccountCreatedBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *CreateAccountCreatedBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *CreateAccountCreatedBody) UnmarshalBinary(b []byte) error {
	var res CreateAccountCreatedBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
