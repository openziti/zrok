// Code generated by go-swagger; DO NOT EDIT.

package rest_model_zrok

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// EnableResponse enable response
//
// swagger:model enableResponse
type EnableResponse struct {

	// cfg
	Cfg string `json:"cfg,omitempty"`

	// identity
	Identity string `json:"identity,omitempty"`
}

// Validate validates this enable response
func (m *EnableResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this enable response based on context it is used
func (m *EnableResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *EnableResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *EnableResponse) UnmarshalBinary(b []byte) error {
	var res EnableResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
