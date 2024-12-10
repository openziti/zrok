// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ListOrganizationsReader is a Reader for the ListOrganizations structure.
type ListOrganizationsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListOrganizationsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListOrganizationsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewListOrganizationsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListOrganizationsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /organizations] listOrganizations", response, response.Code())
	}
}

// NewListOrganizationsOK creates a ListOrganizationsOK with default headers values
func NewListOrganizationsOK() *ListOrganizationsOK {
	return &ListOrganizationsOK{}
}

/*
ListOrganizationsOK describes a response with status code 200, with default header values.

ok
*/
type ListOrganizationsOK struct {
	Payload *ListOrganizationsOKBody
}

// IsSuccess returns true when this list organizations o k response has a 2xx status code
func (o *ListOrganizationsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list organizations o k response has a 3xx status code
func (o *ListOrganizationsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list organizations o k response has a 4xx status code
func (o *ListOrganizationsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list organizations o k response has a 5xx status code
func (o *ListOrganizationsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list organizations o k response a status code equal to that given
func (o *ListOrganizationsOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the list organizations o k response
func (o *ListOrganizationsOK) Code() int {
	return 200
}

func (o *ListOrganizationsOK) Error() string {
	return fmt.Sprintf("[GET /organizations][%d] listOrganizationsOK  %+v", 200, o.Payload)
}

func (o *ListOrganizationsOK) String() string {
	return fmt.Sprintf("[GET /organizations][%d] listOrganizationsOK  %+v", 200, o.Payload)
}

func (o *ListOrganizationsOK) GetPayload() *ListOrganizationsOKBody {
	return o.Payload
}

func (o *ListOrganizationsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(ListOrganizationsOKBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListOrganizationsUnauthorized creates a ListOrganizationsUnauthorized with default headers values
func NewListOrganizationsUnauthorized() *ListOrganizationsUnauthorized {
	return &ListOrganizationsUnauthorized{}
}

/*
ListOrganizationsUnauthorized describes a response with status code 401, with default header values.

unauthorized
*/
type ListOrganizationsUnauthorized struct {
}

// IsSuccess returns true when this list organizations unauthorized response has a 2xx status code
func (o *ListOrganizationsUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list organizations unauthorized response has a 3xx status code
func (o *ListOrganizationsUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list organizations unauthorized response has a 4xx status code
func (o *ListOrganizationsUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list organizations unauthorized response has a 5xx status code
func (o *ListOrganizationsUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list organizations unauthorized response a status code equal to that given
func (o *ListOrganizationsUnauthorized) IsCode(code int) bool {
	return code == 401
}

// Code gets the status code for the list organizations unauthorized response
func (o *ListOrganizationsUnauthorized) Code() int {
	return 401
}

func (o *ListOrganizationsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /organizations][%d] listOrganizationsUnauthorized ", 401)
}

func (o *ListOrganizationsUnauthorized) String() string {
	return fmt.Sprintf("[GET /organizations][%d] listOrganizationsUnauthorized ", 401)
}

func (o *ListOrganizationsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewListOrganizationsInternalServerError creates a ListOrganizationsInternalServerError with default headers values
func NewListOrganizationsInternalServerError() *ListOrganizationsInternalServerError {
	return &ListOrganizationsInternalServerError{}
}

/*
ListOrganizationsInternalServerError describes a response with status code 500, with default header values.

internal server error
*/
type ListOrganizationsInternalServerError struct {
}

// IsSuccess returns true when this list organizations internal server error response has a 2xx status code
func (o *ListOrganizationsInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list organizations internal server error response has a 3xx status code
func (o *ListOrganizationsInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list organizations internal server error response has a 4xx status code
func (o *ListOrganizationsInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this list organizations internal server error response has a 5xx status code
func (o *ListOrganizationsInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this list organizations internal server error response a status code equal to that given
func (o *ListOrganizationsInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the list organizations internal server error response
func (o *ListOrganizationsInternalServerError) Code() int {
	return 500
}

func (o *ListOrganizationsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /organizations][%d] listOrganizationsInternalServerError ", 500)
}

func (o *ListOrganizationsInternalServerError) String() string {
	return fmt.Sprintf("[GET /organizations][%d] listOrganizationsInternalServerError ", 500)
}

func (o *ListOrganizationsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

/*
ListOrganizationsOKBody list organizations o k body
swagger:model ListOrganizationsOKBody
*/
type ListOrganizationsOKBody struct {

	// organizations
	Organizations []*ListOrganizationsOKBodyOrganizationsItems0 `json:"organizations"`
}

// Validate validates this list organizations o k body
func (o *ListOrganizationsOKBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateOrganizations(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *ListOrganizationsOKBody) validateOrganizations(formats strfmt.Registry) error {
	if swag.IsZero(o.Organizations) { // not required
		return nil
	}

	for i := 0; i < len(o.Organizations); i++ {
		if swag.IsZero(o.Organizations[i]) { // not required
			continue
		}

		if o.Organizations[i] != nil {
			if err := o.Organizations[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("listOrganizationsOK" + "." + "organizations" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("listOrganizationsOK" + "." + "organizations" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this list organizations o k body based on the context it is used
func (o *ListOrganizationsOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := o.contextValidateOrganizations(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *ListOrganizationsOKBody) contextValidateOrganizations(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(o.Organizations); i++ {

		if o.Organizations[i] != nil {

			if swag.IsZero(o.Organizations[i]) { // not required
				return nil
			}

			if err := o.Organizations[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("listOrganizationsOK" + "." + "organizations" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("listOrganizationsOK" + "." + "organizations" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (o *ListOrganizationsOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *ListOrganizationsOKBody) UnmarshalBinary(b []byte) error {
	var res ListOrganizationsOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*
ListOrganizationsOKBodyOrganizationsItems0 list organizations o k body organizations items0
swagger:model ListOrganizationsOKBodyOrganizationsItems0
*/
type ListOrganizationsOKBodyOrganizationsItems0 struct {

	// description
	Description string `json:"description,omitempty"`

	// token
	Token string `json:"token,omitempty"`
}

// Validate validates this list organizations o k body organizations items0
func (o *ListOrganizationsOKBodyOrganizationsItems0) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this list organizations o k body organizations items0 based on context it is used
func (o *ListOrganizationsOKBodyOrganizationsItems0) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *ListOrganizationsOKBodyOrganizationsItems0) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *ListOrganizationsOKBodyOrganizationsItems0) UnmarshalBinary(b []byte) error {
	var res ListOrganizationsOKBodyOrganizationsItems0
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}