package dynamicProxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

func TestShareHandlerOptionsBypassesFrontendAuth(t *testing.T) {
	testCases := []struct {
		name       string
		serviceCfg map[string]interface{}
		cfg        *config
		wantFilter bool
	}{
		{
			name: "oauth",
			serviceCfg: map[string]interface{}{
				"auth_scheme": string(sdk.Oauth),
				"oauth": map[string]interface{}{
					"provider":                     "github",
					"authorization_check_interval": "3h",
				},
			},
			cfg: &config{
				Oauth: &oauthConfig{
					CookieName: "zrok_session",
				},
			},
			wantFilter: true,
		},
		{
			name: "basic",
			serviceCfg: map[string]interface{}{
				"auth_scheme": string(sdk.Basic),
				"basic_auth": map[string]interface{}{
					"users": []interface{}{
						map[string]interface{}{
							"username": "demo",
							"password": "secret",
						},
					},
				},
			},
			cfg:        &config{},
			wantFilter: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			restore := stubDynamicProxyService(t, tc.serviceCfg)
			defer restore()

			mappings := newMappings()
			mappings.nameMap["demo.example.com"] = &dynamicProxyController.FrontendMapping{
				Name:       "demo.example.com",
				ShareToken: "share-token",
			}

			upstreamCalled := false
			handler := shareHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upstreamCalled = true
				if got := r.Header.Get("Cookie"); tc.wantFilter && strings.Contains(got, "zrok_session=") {
					t.Fatalf("expected session cookie to be filtered from proxied request, got %q", got)
				}
				w.WriteHeader(http.StatusNoContent)
			}), tc.cfg, nil, nil, mappings)

			req := httptest.NewRequest(http.MethodOptions, "http://demo.example.com/api", nil)
			req.Host = "demo.example.com"
			req.AddCookie(&http.Cookie{Name: "zrok_session", Value: "secret"})

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if !upstreamCalled {
				t.Fatal("expected OPTIONS request to reach upstream handler")
			}
			if rec.Code != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
			}
			if location := rec.Header().Get("Location"); location != "" {
				t.Fatalf("expected no redirect location, got %q", location)
			}
		})
	}
}

func TestShareHandlerGetStillRedirectsForOAuth(t *testing.T) {
	restore := stubDynamicProxyService(t, map[string]interface{}{
		"auth_scheme": string(sdk.Oauth),
		"oauth": map[string]interface{}{
			"provider":                     "github",
			"authorization_check_interval": "3h",
		},
	})
	defer restore()

	mappings := newMappings()
	mappings.nameMap["demo.example.com"] = &dynamicProxyController.FrontendMapping{
		Name:       "demo.example.com",
		ShareToken: "share-token",
	}

	upstreamCalled := false
	handler := shareHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusNoContent)
	}), &config{
		Oauth: &oauthConfig{
			EndpointUrl: "https://oauth.example.com",
			CookieName:  "zrok_session",
		},
	}, nil, nil, mappings)

	req := httptest.NewRequest(http.MethodGet, "http://demo.example.com/api", nil)
	req.Host = "demo.example.com"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if upstreamCalled {
		t.Fatal("expected GET request without session to be blocked by frontend auth")
	}
	if rec.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rec.Code)
	}
	if location := rec.Header().Get("Location"); !strings.Contains(location, "/github/login") {
		t.Fatalf("expected oauth redirect location, got %q", location)
	}
}

func TestShareHandlerGetStillChallengesBasicAuth(t *testing.T) {
	restore := stubDynamicProxyService(t, map[string]interface{}{
		"auth_scheme": string(sdk.Basic),
		"basic_auth": map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{
					"username": "demo",
					"password": "secret",
				},
			},
		},
	})
	defer restore()

	mappings := newMappings()
	mappings.nameMap["demo.example.com"] = &dynamicProxyController.FrontendMapping{
		Name:       "demo.example.com",
		ShareToken: "share-token",
	}

	upstreamCalled := false
	handler := shareHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusNoContent)
	}), &config{}, nil, nil, mappings)

	req := httptest.NewRequest(http.MethodGet, "http://demo.example.com/api", nil)
	req.Host = "demo.example.com"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if upstreamCalled {
		t.Fatal("expected GET request without basic auth credentials to be blocked")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func stubDynamicProxyService(t *testing.T, proxyConfig map[string]interface{}) func() {
	t.Helper()

	previous := getRefreshedService
	getRefreshedService = func(_ string, _ ziti.Context) (*rest_model.ServiceDetail, bool) {
		return &rest_model.ServiceDetail{
			Config: map[string]map[string]interface{}{
				sdk.ZrokProxyConfig: proxyConfig,
			},
		}, true
	}

	return func() {
		getRefreshedService = previous
	}
}
