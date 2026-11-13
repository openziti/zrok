package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/openziti/zrok/v2/environment/env_v0_4"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/stretchr/testify/require"
)

// basicAuthSentinel is a distinctive password; the no-quote tests assert it never appears in an error.
const basicAuthSentinel = "zrokSentinelPw7Q"

// basicAuthShareServer is a fake controller that records the body of every share request.
func basicAuthShareServer(t *testing.T) (*shareHTTPRoot, func() []rest_model_zrok.ShareRequest) {
	var mu sync.Mutex
	var bodies []rest_model_zrok.ShareRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.URL.Path == "/api/v2/share" {
			var body rest_model_zrok.ShareRequest
			require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
			mu.Lock()
			bodies = append(bodies, body)
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"shareToken":"created"}`))
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("ZROK2_API_ENDPOINT", server.URL)
	base, err := env_v0_4.Default()
	require.NoError(t, err)
	return &shareHTTPRoot{Root: base}, func() []rest_model_zrok.ShareRequest {
		mu.Lock()
		defer mu.Unlock()
		return append([]rest_model_zrok.ShareRequest(nil), bodies...)
	}
}

func TestCreateShareSplitsBasicAuthAtFirstColon(t *testing.T) {
	root, bodies := basicAuthShareServer(t)
	_, err := CreateShare(root, &ShareRequest{ShareMode: PublicShareMode, BasicAuth: []string{"admin:s3:cret", " user2 : pass2 "}})
	require.NoError(t, err)
	sent := bodies()
	require.Len(t, sent, 1)
	require.Equal(t, string(Basic), sent[0].AuthScheme)
	require.Len(t, sent[0].BasicAuthUsers, 2)
	require.Equal(t, rest_model_zrok.AuthUser{Username: "admin", Password: "s3:cret"}, *sent[0].BasicAuthUsers[0])
	require.Equal(t, rest_model_zrok.AuthUser{Username: "user2", Password: "pass2"}, *sent[0].BasicAuthUsers[1])
}

func TestCreateShareMalformedBasicAuthDoesNotQuote(t *testing.T) {
	root, bodies := basicAuthShareServer(t)
	_, err := CreateShare(root, &ShareRequest{ShareMode: PublicShareMode, BasicAuth: []string{"user1:pass1", basicAuthSentinel}})
	require.Error(t, err)
	require.NotContains(t, err.Error(), basicAuthSentinel)
	require.Contains(t, err.Error(), "position '2'")
	require.Empty(t, bodies())
}
