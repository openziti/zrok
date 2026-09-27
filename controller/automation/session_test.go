package automation

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
)

// session tests are sequential because they replace the process-wide CA loader.
func sessionFixture(t *testing.T) (*zitifake.Server, *Config, *int) {
	t.Helper()
	fake := zitifake.NewWithCredentials("session-user", "session-password")
	t.Cleanup(fake.Close)
	oldPool := sharedSessions.caPool
	oldClients := sharedSessions.clients
	sharedSessions.clients = make(map[sessionKey]*ZitiAutomation)
	caFetches := 0
	sharedSessions.caPool = func(string) (*x509.CertPool, error) {
		caFetches++
		return nil, nil
	}
	before := zitiAuthentications.Value()
	t.Cleanup(func() {
		successful, _, _ := fake.AuthCounts()
		if got := zitiAuthentications.Value() - before; got != int64(successful) {
			t.Errorf("authentication counter delta = %d, fake = %d", got, successful)
		}
		sharedSessions.caPool = oldPool
		sharedSessions.clients = oldClients
	})
	return fake, &Config{ApiEndpoint: fake.URL, Username: "session-user", Password: "session-password"}, &caFetches
}

func mustSession(t *testing.T, cfg *Config) *ZitiAutomation {
	t.Helper()
	ziti, err := NewZitiAutomation(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return ziti
}

func TestZitiSessionReuse(t *testing.T) {
	fake, cfg, caFetches := sessionFixture(t)
	first := mustSession(t, cfg)
	var id string
	for i := 0; i < 50; i++ {
		ziti := mustSession(t, cfg)
		if ziti != first || ziti.Edge() != first.Edge() {
			t.Fatal("factory did not reuse the client")
		}
		var err error
		switch i % 3 {
		case 0:
			id, err = ziti.Services.Create(&ServiceOptions{BaseOptions: BaseOptions{Name: fmt.Sprintf("reuse-%d", i)}})
		case 1:
			_, err = ziti.Services.GetByID(id)
		case 2:
			err = ziti.Services.Delete(id)
		}
		if err != nil {
			t.Fatalf("operation %d: %v", i, err)
		}
	}
	if successful, attempts, _ := fake.AuthCounts(); successful != 1 || attempts != 1 || *caFetches != 1 {
		t.Fatalf("successful=%d attempts=%d CA fetches=%d", successful, attempts, *caFetches)
	}
}

func TestZitiSessionConcurrentExpiry(t *testing.T) {
	fake, cfg, caFetches := sessionFixture(t)
	ziti := mustSession(t, cfg)
	fake.ExpireAll()
	start := make(chan struct{})
	failures := make(chan error, 20)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() {
			<-start
			client, err := NewZitiAutomation(cfg)
			if err == nil && client != ziti {
				err = errors.New("factory did not reuse the client")
			}
			if err == nil {
				if i%2 == 0 {
					_, err = client.Services.Create(&ServiceOptions{BaseOptions: BaseOptions{Name: fmt.Sprintf("concurrent-%d", i)}})
				} else {
					_, err = client.Services.Find(&FilterOptions{})
				}
			}
			failures <- err
		})
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Error(err)
		}
	}
	successful, attempts, unauthorized := fake.AuthCounts()
	if refreshes := successful - 1; refreshes < 1 || refreshes > 3 || attempts != successful || unauthorized < 1 {
		t.Fatalf("successful=%d attempts=%d unauthorized=%d", successful, attempts, unauthorized)
	}
	if *caFetches != 1 {
		t.Fatalf("CA fetched %d times", *caFetches)
	}
	_, _, creates, _ := fake.Counts()
	if creates != 10 {
		t.Fatalf("created %d services, want 10", creates)
	}
}

func TestZitiSessionReplayBody(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	ziti := mustSession(t, cfg)
	fake.ExpireAll()
	maxIdle := int64(1234)
	opts := &ServiceOptions{
		BaseOptions: BaseOptions{Name: "replayed-service"},
		Configs:     []string{"config-one", "config-two"}, EncryptionRequired: true,
		TerminatorStrategy: "smartrouting", RoleAttributes: []string{"role-one", "role-two"}, MaxIdleTime: &maxIdle,
	}
	id, err := ziti.Services.Create(opts)
	if err != nil {
		t.Fatal(err)
	}
	created, err := ziti.Services.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if *created.Name != opts.Name || !*created.EncryptionRequired || !reflect.DeepEqual(created.Configs, opts.Configs) || !reflect.DeepEqual([]string(*created.RoleAttributes), opts.RoleAttributes) || *created.TerminatorStrategy != opts.TerminatorStrategy || *created.MaxIdleTimeMillis != maxIdle {
		t.Fatalf("replayed service lost fields: %+v", created)
	}
	if successful, attempts, unauthorized := fake.AuthCounts(); successful != 2 || attempts != 2 || unauthorized != 1 {
		t.Fatalf("successful=%d attempts=%d unauthorized=%d", successful, attempts, unauthorized)
	}
}

func TestZitiSessionPersistentUnauthorized(t *testing.T) {
	for _, rejectLogin := range []bool{true, false} {
		t.Run(fmt.Sprintf("reject-login-%t", rejectLogin), func(t *testing.T) {
			fake, cfg, _ := sessionFixture(t)
			ziti := mustSession(t, cfg)
			fake.ExpireAll()
			fake.RejectAuthentication(rejectLogin)
			fake.RejectOperations(!rejectLogin)
			done := make(chan error, 1)
			go func() {
				_, err := ziti.Services.Find(&FilterOptions{Timeout: time.Second})
				done <- err
			}()
			select {
			case err := <-done:
				var unauthorized *service.ListServicesUnauthorized
				if !errors.As(err, &unauthorized) {
					t.Fatalf("want typed unauthorized, got %T: %v", err, err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("operation did not finish within its retry budget")
			}
			successful, attempts, unauthorized := fake.AuthCounts()
			wantSuccess, wantUnauthorized := 2, 2
			if rejectLogin {
				wantSuccess, wantUnauthorized = 1, 1
			}
			if successful != wantSuccess || attempts != 2 || unauthorized != wantUnauthorized {
				t.Fatalf("successful=%d attempts=%d unauthorized=%d", successful, attempts, unauthorized)
			}
		})
	}
}

func TestZitiSessionFailedBuildNotCached(t *testing.T) {
	for _, failure := range []string{"down", "credentials", "ca"} {
		t.Run(failure, func(t *testing.T) {
			fake, cfg, caFetches := sessionFixture(t)
			var recoverServer func()
			switch failure {
			case "down":
				server := httptest.NewUnstartedServer(fake.Config.Handler)
				// close the server before the first build, then recover on the same port.
				cfg.ApiEndpoint = "http://" + server.Listener.Addr().String()
				server.Start()
				server.Close()
				recoverServer = func() {
					// restart on the same endpoint to exercise the same cache key.
					restarted := httptest.NewUnstartedServer(fake.Config.Handler)
					restarted.Listener.Close()
					listener, err := net.Listen("tcp", server.Listener.Addr().String())
					if err != nil {
						t.Fatal(err)
					}
					restarted.Listener = listener
					restarted.Start()
					t.Cleanup(restarted.Close)
				}
			case "credentials":
				cfg.Password = "wrong-password"
				recoverServer = func() { cfg.Password = "session-password" }
			case "ca":
				sharedSessions.caPool = func(string) (*x509.CertPool, error) {
					*caFetches++
					if *caFetches == 1 {
						return nil, errors.New("CA fetch failed")
					}
					return nil, nil
				}
				recoverServer = func() {}
			}
			if ziti, err := NewZitiAutomation(cfg); err == nil || ziti != nil {
				t.Fatalf("failed build returned client=%v error=%v", ziti, err)
			}
			recoverServer()
			ziti := mustSession(t, cfg)
			if _, err := ziti.Services.Find(&FilterOptions{}); err != nil {
				t.Fatal(err)
			}
			if *caFetches != 2 {
				t.Fatalf("CA fetched %d times, want 2", *caFetches)
			}
		})
	}
}

func TestZitiSessionCounterEndpoint(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	ziti := mustSession(t, cfg)
	fake.ExpireAll()
	if _, err := ziti.Services.Find(&FilterOptions{}); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/debug/vars", nil))
	var vars map[string]json.RawMessage
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &vars) != nil {
		t.Fatalf("invalid /debug/vars response: %s", response.Body)
	}
	if got := string(vars["zrok.ziti.authentications"]); got != expvar.Get("zrok.ziti.authentications").String() {
		t.Fatalf("counter endpoint = %s, want %s", got, zitiAuthentications.String())
	}
}
