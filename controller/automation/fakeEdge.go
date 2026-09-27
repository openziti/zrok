package automation

import "github.com/openziti/edge-api/rest_management_api_client"

func NewZitiAutomationWithEdge(edge *rest_management_api_client.ZitiEdgeManagement) *ZitiAutomation {
	ziti := &ZitiAutomation{edge: edge}
	ziti.Identities = NewIdentityManager(ziti)
	ziti.Services = NewServiceManager(ziti)
	ziti.Configs = NewConfigManager(ziti)
	ziti.ConfigTypes = NewConfigTypeManager(ziti)
	ziti.EdgeRouterPolicies = NewEdgeRouterPolicyManager(ziti)
	ziti.ServiceEdgeRouterPolicies = NewServiceEdgeRouterPolicyManager(ziti)
	ziti.ServicePolicies = NewServicePolicyManager(ziti)
	return ziti
}
