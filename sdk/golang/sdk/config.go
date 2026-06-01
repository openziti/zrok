package sdk

import (
	"github.com/pkg/errors"
	"reflect"
)

const ZrokProxyConfig = "zrok.proxy.v1"

type FrontendConfig struct {
	Interstitial bool             `json:"interstitial"`
	AuthScheme   AuthScheme       `json:"auth_scheme"`
	BasicAuth    *BasicAuthConfig `json:"basic_auth"`
	OauthAuth    *OauthConfig     `json:"oauth"`
}

func FrontendConfigFromMap(m map[string]interface{}) (*FrontendConfig, error) {
	out := &FrontendConfig{}
	if v, found := m["interstitial"]; found {
		out.Interstitial = v.(bool)
	}
	if v, found := m["auth_scheme"]; found {
		if vStr, ok := v.(string); ok {
			out.AuthScheme = AuthScheme(vStr)
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	if v, found := m["basic_auth"]; found && v != nil {
		if subMap, ok := v.(map[string]interface{}); ok {
			ba, err := BasicAuthConfigFromMap(subMap)
			if err != nil {
				return nil, err
			}
			out.BasicAuth = ba
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	if v, found := m["oauth"]; found && v != nil {
		if subMap, ok := v.(map[string]interface{}); ok {
			o, err := OauthConfigFromMap(subMap)
			if err != nil {
				return nil, err
			}
			out.OauthAuth = o
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	return out, nil
}

type BasicAuthConfig struct {
	Users []*AuthUserConfig `json:"users"`
}

func BasicAuthConfigFromMap(m map[string]interface{}) (*BasicAuthConfig, error) {
	out := &BasicAuthConfig{}
	if v, found := m["users"]; found {
		if subArr, ok := v.([]interface{}); ok {
			for _, v := range subArr {
				if subMap, ok := v.(map[string]interface{}); ok {
					if auc, err := AuthUserConfigFromMap(subMap); err == nil {
						out.Users = append(out.Users, auc)
					} else {
						return nil, err
					}
				} else {
					return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
				}
			}
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	} else {
		return nil, errors.New("missing 'users' field")
	}
	return out, nil
}

type AuthUserConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func AuthUserConfigFromMap(m map[string]interface{}) (*AuthUserConfig, error) {
	auc := &AuthUserConfig{}
	if v, found := m["username"]; found {
		if vStr, ok := v.(string); ok {
			auc.Username = vStr
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	if v, found := m["password"]; found {
		if vStr, ok := v.(string); ok {
			auc.Password = vStr
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	return auc, nil
}

type OauthConfig struct {
	Provider                   string   `json:"provider"`
	EmailDomains               []string `json:"email_domains"`
	AuthorizationCheckInterval string   `json:"authorization_check_interval"`
}

func OauthConfigFromMap(m map[string]interface{}) (*OauthConfig, error) {
	oac := &OauthConfig{}
	if v, found := m["provider"]; found {
		if vStr, ok := v.(string); ok {
			oac.Provider = vStr
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	if v, found := m["email_domains"]; found {
		if vArr, ok := v.([]interface{}); ok {
			for _, vV := range vArr {
				if vStr, ok := vV.(string); ok {
					oac.EmailDomains = append(oac.EmailDomains, vStr)
				} else {
					return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(vV))
				}
			}
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	if v, found := m["authorization_check_interval"]; found {
		if vStr, ok := v.(string); ok {
			oac.AuthorizationCheckInterval = vStr
		} else {
			return nil, errors.Errorf("unexpected type '%v'", reflect.TypeOf(v))
		}
	}
	return oac, nil
}

func ParseAuthScheme(authScheme string) (AuthScheme, error) {
	switch authScheme {
	case string(None):
		return None, nil
	case string(Basic):
		return Basic, nil
	case string(Oauth):
		return Oauth, nil
	default:
		return None, errors.Errorf("unknown auth scheme '%v'", authScheme)
	}
}
