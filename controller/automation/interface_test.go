package automation

import (
	"testing"

	"github.com/openziti/edge-api/rest_model"
)

// test that demonstrates the unified interface functionality
func TestUnifiedInterface(t *testing.T) {
	// test the generic helper functions work with all resource types

	// test GetByID with mock data
	mockFinder := func(opts *FilterOptions) ([]*rest_model.IdentityDetail, error) {
		if opts.Filter == `id="id-1"` {
			return []*rest_model.IdentityDetail{{}}, nil
		}
		return []*rest_model.IdentityDetail{}, nil
	}

	identity, err := GetByID(mockFinder, "id-1", "identity")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if identity == nil {
		t.Fatal("expected identity to be returned")
	}

	// test not found case
	_, err = GetByID(mockFinder, "id-nonexistent", "identity")
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}

	if automationErr, ok := err.(*AutomationError); ok {
		if !automationErr.IsNotFound() {
			t.Fatal("expected not found error")
		}
	}
}

// test that all managers implement the interface
func TestManagerInterfaces(t *testing.T) {
	ziti := &ZitiAutomation{} // mock client for testing

	// verify all managers implement their interfaces
	var _ IResourceManager[rest_model.IdentityDetail, *IdentityOptions] = NewIdentityManager(ziti)
	var _ IResourceManager[rest_model.ServiceDetail, *ServiceOptions] = NewServiceManager(ziti)
	var _ IResourceManager[rest_model.ConfigDetail, *ConfigOptions] = NewConfigManager(ziti)
	var _ IResourceManager[rest_model.ConfigTypeDetail, *ConfigTypeOptions] = NewConfigTypeManager(ziti)
}

// test error types
func TestErrorTypes(t *testing.T) {
	err := NewNotFoundError("identity", "GetByID", nil)

	if !err.IsNotFound() {
		t.Fatal("expected IsNotFound to return true")
	}

	if err.IsRetryable() {
		t.Fatal("expected IsRetryable to return false for not found error")
	}
}

func stringPtr(s string) *string {
	return &s
}
