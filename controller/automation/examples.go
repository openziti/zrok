package automation

import (
	"github.com/openziti/edge-api/rest_model"
)

// Examples demonstrating the zrok automation framework
//
// Key improvements in the unified interface:
// 1. Consistent CRUD operations across all resource types
// 2. Flattened options pattern (no nested ResourceOptions)
// 3. Generic helper functions eliminate code duplication
// 4. Structured error handling with AutomationError
// 5. Type-safe operations with compile-time guarantees

// Basic Resource Operations
//
// All managers (Identities, Services, Configs, ConfigTypes) implement the same interface:
//   - Create(opts) (string, error)
//   - Delete(id string) error
//   - Find(opts *FilterOptions) ([]*T, error)
//   - GetByID(id string) (*T, error)
//   - GetByName(name string) (*T, error)

func ExampleBasicOperations(za *ZitiAutomation) {
	// consistent interface across all resource types
	identity, err := za.Identities.GetByID("identity-id")
	service, err := za.Services.GetByName("service-name")
	config, err := za.Configs.GetByID("config-id")
	
	// standardized error handling
	if automationErr, ok := err.(*AutomationError); ok {
		if automationErr.IsNotFound() {
			// handle missing resource
		}
		if automationErr.IsRetryable() {
			// retry network errors
		}
	}
	
	_, _, _ = identity, service, config
}

// Creating Resources with Flattened Options
//
// Before: nested ResourceOptions caused cognitive overhead
// After: direct field access in flat structure

func ExampleCreateIdentity(za *ZitiAutomation) (string, error) {
	tags := NewZrokTagStrategy().WithTag("zrokEmail", "user@example.com")
	
	opts := &IdentityOptions{
		BaseOptions: BaseOptions{
			Name: "user-identity",
			Tags: tags,
		},
		Type:    rest_model.IdentityTypeUser,
		IsAdmin: false,
	}
	
	return za.Identities.Create(opts)
}

func ExampleCreateService(za *ZitiAutomation, configID string) (string, error) {
	tags := ZrokBaseTags()
	
	opts := &ServiceOptions{
		BaseOptions: BaseOptions{
			Name: "my-service",
			Tags: tags,
		},
		Configs:            []string{configID},
		EncryptionRequired: true,
	}
	
	return za.Services.Create(opts)
}

func ExampleCreateConfig(za *ZitiAutomation, configTypeID string) (string, error) {
	tags := ZrokShareTags("share-token-123")
	
	opts := &ConfigOptions{
		BaseOptions: BaseOptions{
			Name: "proxy-config",
			Tags: tags,
		},
		ConfigTypeID: configTypeID,
		Data:         map[string]interface{}{"address": "localhost:8080"},
	}
	
	return za.Configs.Create(opts)
}

// High-Level Workflow Examples
//
// Common zrok patterns using the streamlined API

func ExampleCreateEnvironment(za *ZitiAutomation, accountEmail, uniqueToken, envDescription string) (string, error) {
	name := accountEmail + "-" + uniqueToken + "-" + envDescription
	tags := NewZrokTagStrategy().WithTag("zrokEmail", accountEmail)
	
	opts := &IdentityOptions{
		BaseOptions: BaseOptions{
			Name: name,
			Tags: tags,
		},
		Type:    rest_model.IdentityTypeUser,
		IsAdmin: false,
	}
	
	return za.Identities.Create(opts)
}

func ExampleCreateShare(za *ZitiAutomation, envZId, shareToken, configID string) (string, error) {
	tags := ZrokShareTags(shareToken)
	
	// create service using convenience method
	serviceID, err := za.CreateServiceWithConfig(shareToken, configID, tags)
	if err != nil {
		return "", err
	}
	
	// create bind policy
	bindName := shareToken + "-bind"
	_, err = za.CreateBindPolicy(bindName, serviceID, envZId, tags)
	if err != nil {
		return "", err
	}
	
	// create edge router policy for service
	serpName := shareToken + "-serp"
	_, err = za.CreateServiceEdgeRouterPolicy(serpName, serviceID, tags)
	if err != nil {
		return "", err
	}
	
	return serviceID, nil
}

func ExampleCreateAccess(za *ZitiAutomation, envZId, serviceID string, dialIdentities []string) error {
	tags := ZrokBaseTags()
	dialName := serviceID + "-dial"
	_, err := za.CreateDialPolicy(dialName, serviceID, dialIdentities, tags)
	return err
}

// Bulk Operations and Cleanup
//
// Generic functions work consistently across all resource types

func ExampleBulkOperations(za *ZitiAutomation) error {
	// find resources with filter
	filter := "tags.environment = 'test'"
	identities, err := za.Identities.Find(&FilterOptions{Filter: filter})
	if err != nil {
		return err
	}
	
	// bulk delete using generic helper (available on all managers)
	err = za.Identities.DeleteWithFilter("tags.cleanup = true")
	if err != nil {
		return err
	}
	
	// same pattern works for all resource types
	err = za.Services.DeleteWithFilter("tags.temporary = true")
	if err != nil {
		return err
	}
	
	_ = identities
	return nil
}

func ExampleCleanupShare(za *ZitiAutomation, shareToken string) error {
	return za.CleanupByTagFilter("zrokShareToken", shareToken)
}

func ExampleCleanupEnvironment(za *ZitiAutomation, envZId string) error {
	return za.CleanupByTagFilter("zrokEnvZId", envZId)
}

// Tag Strategies
//
// Flexible tagging system with pre-built strategies for common patterns

func ExampleTagStrategies() {
	// simple key-value tags
	simpleTags := NewSimpleTagStrategy(map[string]interface{}{
		"environment": "production",
		"team":        "backend",
	})
	
	// zrok-specific tags with automatic build info
	zrokTags := NewZrokTagStrategy().
		WithTag("zrokEmail", "user@example.com").
		WithShareToken("share-123")
	
	// convenience functions for common patterns
	shareTags := ZrokShareTags("share-token")
	agentTags := ZrokAgentRemoteTags("enrollment-token", "env-zid")
	baseTags := ZrokBaseTags() // includes build info
	
	_, _, _, _, _ = simpleTags, zrokTags, shareTags, agentTags, baseTags
}

// Advanced Error Handling
//
// Structured errors with type information and retry logic

func ExampleErrorHandling(za *ZitiAutomation) {
	identity, err := za.Identities.GetByID("non-existent")
	if err != nil {
		if automationErr, ok := err.(*AutomationError); ok {
			switch automationErr.Type {
			case ErrorTypeNotFound:
				// resource doesn't exist
			case ErrorTypePermission:
				// insufficient permissions
			case ErrorTypeNetwork:
				// network issue - can retry
				if automationErr.IsRetryable() {
					// implement retry logic
				}
			case ErrorTypeValidation:
				// invalid input data
			case ErrorTypeInternal:
				// internal server error
			}
		}
	}
	_ = identity
}