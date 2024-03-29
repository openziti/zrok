// Code generated by go-swagger; DO NOT EDIT.

package rest_model_zrok

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VerifyRequest verify request
//
// swagger:model verifyRequest
type VerifyRequest struct {

	// token
	Token string `json:"token,omitempty"`
}

// Validate validates this verify request
func (m *VerifyRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this verify request based on context it is used
func (m *VerifyRequest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VerifyRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VerifyRequest) UnmarshalBinary(b []byte) error {
	var res VerifyRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
