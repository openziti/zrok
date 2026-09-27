package zrokEdgeSdk

import (
	"github.com/openziti/edge-api/rest_management_api_client"
)

type Config struct {
	ApiEndpoint string
	Username    string
	Password    string `cf:"+secret"`
}

func Client(cfg *Config) (*rest_management_api_client.ZitiEdgeManagement, error) {
	return sharedSessions.get(cfg)
}
