package zrokEdgeSdk

import (
	"context"
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
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	edge_service "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/zrokEdgeSdk/zitifake"
)

// session tests are sequential because they replace the process-wide CA loader. extra fakes serving
// other cache keys are included in the authentication counter check.
func sessionFixture(t *testing.T, extra ...*zitifake.Server) (*zitifake.Server, *Config, *atomic.Int64) {
	t.Helper()
	fake := zitifake.NewWithCredentials("session-user", "session-password")
	t.Cleanup(fake.Close)
	oldPool := sharedSessions.caPool
	oldClients := sharedSessions.clients
	sharedSessions.clients = make(map[sessionKey]*rest_management_api_client.ZitiEdgeManagement)
	caFetches := &atomic.Int64{}
	sharedSessions.caPool = func(string) (*x509.CertPool, error) {
		caFetches.Add(1)
		return nil, nil
	}
	before := zitiAuthentications.Value()
	t.Cleanup(func() {
		successful := 0
		for _, f := range append([]*zitifake.Server{fake}, extra...) {
			n, _, _ := f.AuthCounts()
			successful += n
		}
		if got := zitiAuthentications.Value() - before; got != int64(successful) {
			t.Errorf("authentication counter delta = %d, fake = %d", got, successful)
		}
		sharedSessions.caPool = oldPool
		sharedSessions.clients = oldClients
	})
	return fake, &Config{ApiEndpoint: fake.URL, Username: "session-user", Password: "session-password"}, caFetches
}

func mustSession(t *testing.T, cfg *Config) *rest_management_api_client.ZitiEdgeManagement {
	t.Helper()
	edge, err := Client(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return edge
}

func listServices(edge *rest_management_api_client.ZitiEdgeManagement, timeout time.Duration) error {
	req := &edge_service.ListServicesParams{Context: context.Background()}
	req.SetTimeout(timeout)
	_, err := edge.Service.ListServices(req, nil)
	return err
}

func TestZitiSessionReuse(t *testing.T) {
	fake, cfg, caFetches := sessionFixture(t)
	first := mustSession(t, cfg)
	var id string
	for i := 0; i < 50; i++ {
		edge := mustSession(t, cfg)
		if edge != first {
			t.Fatal("factory did not reuse the client")
		}
		var err error
		switch i % 3 {
		case 0:
			id, err = CreateService(fmt.Sprintf("reuse-%d", i), nil, nil, edge)
		case 1:
			_, err = FindShareService(id, edge)
		case 2:
			err = DeleteService("env", id, edge)
		}
		if err != nil {
			t.Fatalf("operation %d: %v", i, err)
		}
	}
	if successful, attempts, _ := fake.AuthCounts(); successful != 1 || attempts != 1 || caFetches.Load() != 1 {
		t.Fatalf("successful=%d attempts=%d CA fetches=%d", successful, attempts, caFetches.Load())
	}
}

func TestZitiSessionConcurrentExpiry(t *testing.T) {
	fake, cfg, caFetches := sessionFixture(t)
	edge := mustSession(t, cfg)
	fake.ExpireAll()
	start := make(chan struct{})
	failures := make(chan error, 20)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			client, err := Client(cfg)
			if err == nil && client != edge {
				err = errors.New("factory did not reuse the client")
			}
			if err == nil {
				if i%2 == 0 {
					_, err = CreateService(fmt.Sprintf("concurrent-%d", i), nil, nil, client)
				} else {
					err = listServices(client, 30*time.Second)
				}
			}
			failures <- err
		}()
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
	if caFetches.Load() != 1 {
		t.Fatalf("CA fetched %d times", caFetches.Load())
	}
	_, _, creates, _ := fake.Counts()
	if creates != 10 {
		t.Fatalf("created %d services, want 10", creates)
	}
}

func TestZitiSessionReplayBody(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	edge := mustSession(t, cfg)
	fake.ExpireAll()
	name, encryption, strategy, maxIdle := "replayed-service", true, "smartrouting", int64(1234)
	svc := &rest_model.ServiceCreate{
		Name: &name, EncryptionRequired: &encryption, TerminatorStrategy: strategy, MaxIdleTimeMillis: maxIdle,
		Configs: []string{"config-one", "config-two"}, RoleAttributes: []string{"role-one", "role-two"},
	}
	createReq := &edge_service.CreateServiceParams{Service: svc, Context: context.Background()}
	createReq.SetTimeout(30 * time.Second)
	resp, err := edge.Service.CreateService(createReq, nil)
	if err != nil {
		t.Fatal(err)
	}
	detailReq := &edge_service.DetailServiceParams{ID: resp.Payload.Data.ID, Context: context.Background()}
	detailReq.SetTimeout(30 * time.Second)
	detail, err := edge.Service.DetailService(detailReq, nil)
	if err != nil {
		t.Fatal(err)
	}
	created := detail.Payload.Data
	if *created.Name != name || !*created.EncryptionRequired || !reflect.DeepEqual(created.Configs, svc.Configs) || !reflect.DeepEqual([]string(*created.RoleAttributes), svc.RoleAttributes) || *created.TerminatorStrategy != strategy || *created.MaxIdleTimeMillis != maxIdle {
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
			edge := mustSession(t, cfg)
			fake.ExpireAll()
			fake.RejectAuthentication(rejectLogin)
			fake.RejectOperations(!rejectLogin)
			done := make(chan error, 1)
			go func() {
				done <- listServices(edge, time.Second)
			}()
			select {
			case err := <-done:
				var unauthorized *edge_service.ListServicesUnauthorized
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
					if caFetches.Add(1) == 1 {
						return nil, errors.New("CA fetch failed")
					}
					return nil, nil
				}
				recoverServer = func() {}
			}
			if edge, err := Client(cfg); err == nil || edge != nil {
				t.Fatalf("failed build returned client=%v error=%v", edge, err)
			}
			recoverServer()
			edge := mustSession(t, cfg)
			if err := listServices(edge, 30*time.Second); err != nil {
				t.Fatal(err)
			}
			if caFetches.Load() != 2 {
				t.Fatalf("CA fetched %d times, want 2", caFetches.Load())
			}
		})
	}
}

func TestZitiSessionCounterEndpoint(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	edge := mustSession(t, cfg)
	fake.ExpireAll()
	if err := listServices(edge, 30*time.Second); err != nil {
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

func TestZitiSessionSustainedUnauthorizedBounded(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	edge := mustSession(t, cfg)
	fake.ExpireAll()
	fake.RejectAuthentication(true)
	delay := 500 * time.Millisecond
	fake.SetAuthenticationDelay(delay)
	start := make(chan struct{})
	results := make(chan error, 20)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- listServices(edge, 30*time.Second)
		}()
	}
	began := time.Now()
	close(start)
	wg.Wait()
	elapsed := time.Since(began)
	close(results)
	for err := range results {
		var unauthorized *edge_service.ListServicesUnauthorized
		if !errors.As(err, &unauthorized) {
			t.Errorf("want typed unauthorized, got %T: %v", err, err)
		}
	}
	if elapsed > 2*delay {
		t.Fatalf("twenty operations took %v behind a %v login", elapsed, delay)
	}
	if _, attempts, _ := fake.AuthCounts(); attempts-1 < 1 || attempts-1 > 3 {
		t.Fatalf("re-authentication attempts = %d", attempts-1)
	}
}

func TestZitiSessionFailedRefreshRemembered(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	oldClock := refreshClock
	t.Cleanup(func() { refreshClock = oldClock })
	edge := mustSession(t, cfg)
	fake.ExpireAll()
	fake.RejectAuthentication(true)
	attempts := func() int {
		_, n, _ := fake.AuthCounts()
		return n
	}
	operation := func() error {
		return listServices(edge, 30*time.Second)
	}
	if err := operation(); err == nil || attempts() != 2 {
		t.Fatalf("first operation error=%v attempts=%d", err, attempts())
	}
	if err := operation(); err == nil || attempts() != 2 {
		t.Fatalf("remembered failure error=%v attempts=%d", err, attempts())
	}
	refreshClock = func() time.Time { return time.Now().Add(failedRefreshWindow) }
	if err := operation(); err == nil || attempts() != 3 {
		t.Fatalf("after window error=%v attempts=%d", err, attempts())
	}
}

func TestZitiSessionConcurrentFirstBuild(t *testing.T) {
	other := zitifake.NewWithCredentials("session-user", "session-password")
	t.Cleanup(other.Close)
	fake, cfg, _ := sessionFixture(t, other)
	delay := 500 * time.Millisecond
	fake.SetAuthenticationDelay(delay)
	other.SetAuthenticationDelay(delay)
	clients := make(chan *rest_management_api_client.ZitiEdgeManagement, 10)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			edge, err := Client(cfg)
			if err != nil {
				t.Error(err)
			}
			clients <- edge
		}()
	}
	// let the first key's build start before timing the second key.
	time.Sleep(100 * time.Millisecond)
	began := time.Now()
	if _, err := Client(&Config{ApiEndpoint: other.URL, Username: "session-user", Password: "session-password"}); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(began); elapsed > delay+300*time.Millisecond {
		t.Fatalf("other key waited %v behind a %v login", elapsed, delay)
	}
	wg.Wait()
	close(clients)
	var first *rest_management_api_client.ZitiEdgeManagement
	for edge := range clients {
		if first == nil {
			first = edge
		}
		if edge == nil || edge != first {
			t.Fatal("concurrent builds returned different clients")
		}
	}
	if successful, attempts, _ := fake.AuthCounts(); successful != 1 || attempts != 1 {
		t.Fatalf("successful=%d attempts=%d", successful, attempts)
	}
}
