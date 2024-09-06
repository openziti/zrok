package sdk

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicFrontendConfigFromMap(t *testing.T) {
	inFec := &FrontendConfig{
		Interstitial: true,
		AuthScheme:   None,
	}
	m, err := frontendConfigToMap(inFec)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	outFec, err := FrontendConfigFromMap(m)
	assert.NoError(t, err)
	assert.NotNil(t, outFec)
	assert.Equal(t, inFec, outFec)
}

func TestBasicAuthFrontendConfigFromMap(t *testing.T) {
	inFec := &FrontendConfig{
		Interstitial: false,
		AuthScheme:   Basic,
		BasicAuth: &BasicAuthConfig{
			Users: []*AuthUserConfig{
				{Username: "nobody", Password: "password"},
			},
		},
	}
	m, err := frontendConfigToMap(inFec)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	outFec, err := FrontendConfigFromMap(m)
	assert.NoError(t, err)
	assert.NotNil(t, outFec)
	assert.Equal(t, inFec, outFec)
}

func TestOauthAuthFrontendConfigFromMap(t *testing.T) {
	inFec := &FrontendConfig{
		Interstitial: true,
		AuthScheme:   Oauth,
		OauthAuth: &OauthConfig{
			Provider:                   "google",
			EmailDomains:               []string{"a@b.com", "c@d.com"},
			AuthorizationCheckInterval: "5m",
		},
	}
	m, err := frontendConfigToMap(inFec)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	outFec, err := FrontendConfigFromMap(m)
	assert.NoError(t, err)
	assert.NotNil(t, outFec)
	assert.Equal(t, inFec, outFec)
}

func frontendConfigToMap(fec *FrontendConfig) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(fec)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &m); err != nil {
		return nil, err
	}
	return m, nil
}
