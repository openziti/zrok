// Code generated by go-swagger; DO NOT EDIT.

package rest_model_zrok

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ChangePasswordRequest change password request
//
// swagger:model changePasswordRequest
type ChangePasswordRequest struct {

	// email
	Email string `json:"email,omitempty"`

	// new password
	NewPassword string `json:"newPassword,omitempty"`

	// old password
	OldPassword string `json:"oldPassword,omitempty"`
}

// Validate validates this change password request
func (m *ChangePasswordRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this change password request based on context it is used
func (m *ChangePasswordRequest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ChangePasswordRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ChangePasswordRequest) UnmarshalBinary(b []byte) error {
	var res ChangePasswordRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
