package automation

import (
	"github.com/openziti/edge-api/rest_model"
)

// examples of how to use the new framework with zrok-specific tag strategies

func ExampleCreateEnvironment(za *ZitiAutomation, accountEmail, uniqueToken, envDescription string) (string, error) {
	name := accountEmail + "-" + uniqueToken + "-" + envDescription
	tags := NewZrokTagStrategy().WithTag("zrokEmail", accountEmail)

	identityType := rest_model.IdentityTypeUser
	opts := &IdentityOptions{
		ResourceOptions: &ResourceOptions{
			Name: name,
			Tags: tags,
		},
		Type:    identityType,
		IsAdmin: false,
	}

	return za.Identities.Create(opts)
}

func ExampleCreateShare(za *ZitiAutomation, envZId, shareToken, configID string) (string, error) {
	// create service
	tags := ZrokShareTags(shareToken)

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
	// create dial policy
	tags := ZrokBaseTags()

	dialName := serviceID + "-dial"
	_, err := za.CreateDialPolicy(dialName, serviceID, dialIdentities, tags)
	return err
}

func ExampleCleanupShare(za *ZitiAutomation, shareToken string) error {
	return za.CleanupByTagFilter("zrokShareToken", shareToken)
}

func ExampleCleanupEnvironment(za *ZitiAutomation, envZId string) error {
	return za.CleanupByTagFilter("zrokEnvZId", envZId)
}
