//go:build agent_smoke

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/v2/agent/agentClient"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"google.golang.org/grpc"
)

const (
	fakeCLIEnv   = "AGENT_SMOKE_FAKE_CLI"
	fakeModeEnv  = "AGENT_SMOKE_FAKE_MODE"
	fakeBuildEnv = "AGENT_SMOKE_FAKE_BUILD"
)

// fakeController serves the generated api paths the harness reads and writes, plus fake-only paths the fake
// binary uses to create resources. a second server stands in for the public frontend.
type fakeController struct {
	mu            sync.Mutex
	accountToken  string
	refuseVersion bool
	wrongBody     bool
	failReads     int
	failOnStop    int
	envs          map[string]*fakeEnv
	names         map[string]*fakeName
	shares        map[string]*fakeShare
	calls         []string
	next          int
	api           *httptest.Server
	frontend      *httptest.Server
}

type fakeEnv struct {
	zID, description string
}

type fakeName struct {
	namespace, name, share string
}

type fakeShare struct {
	token, envZID, target, name string
}

func newFakeController(t *testing.T, accountToken string) *fakeController {
	f := &fakeController{
		accountToken: accountToken,
		envs:         make(map[string]*fakeEnv),
		names:        make(map[string]*fakeName),
		shares:       make(map[string]*fakeShare),
	}
	f.api = httptest.NewServer(http.HandlerFunc(f.serveAPI))
	f.frontend = httptest.NewServer(http.HandlerFunc(f.serveFrontend))
	t.Cleanup(f.api.Close)
	t.Cleanup(f.frontend.Close)
	return f
}

func (f *fakeController) callCount(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, call := range f.calls {
		if call == name {
			count++
		}
	}
	return count
}

func (f *fakeController) empty() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.envs) == 0 && len(f.names) == 0 && len(f.shares) == 0
}

func (f *fakeController) addEnv(zID, description string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.envs[zID] = &fakeEnv{zID: zID, description: description}
}

func (f *fakeController) addName(namespace, name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.names[namespace+":"+name] = &fakeName{namespace: namespace, name: name}
}

func (f *fakeController) serveAPI(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path := strings.TrimPrefix(r.URL.Path, "/api/v2")
	f.calls = append(f.calls, r.Method+" "+path)
	var body map[string]any
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	field := func(key string) string {
		value, _ := body[key].(string)
		return value
	}
	if path == "/clientVersionCheck" {
		if f.refuseVersion {
			writeJSON(w, http.StatusBadRequest, fmt.Sprintf("client version '%v' does not match any accepted patterns", field("clientVersion")))
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Header.Get("X-TOKEN") != f.accountToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch r.Method + " " + path {
	case "GET /overview":
		if f.failReads > 0 {
			f.failReads--
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, f.overview())
	case "GET /share/names":
		writeJSON(w, http.StatusOK, f.nameList())
	case "DELETE /unshare":
		shr, found := f.shares[field("shareToken")]
		if !found || shr.envZID != field("envZId") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f.deleteShare(shr.token, false)
		w.WriteHeader(http.StatusOK)
	case "DELETE /share/name":
		key := field("namespaceToken") + ":" + field("name")
		name, found := f.names[key]
		switch {
		case !found:
			w.WriteHeader(http.StatusNotFound)
		case name.share != "" && f.shares[name.share] != nil:
			writeJSON(w, http.StatusConflict, "name is held by a live share")
		default:
			delete(f.names, key)
			w.WriteHeader(http.StatusOK)
		}
	case "POST /disable":
		zID := field("identity")
		if _, found := f.envs[zID]; !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		for token, shr := range f.shares {
			if shr.envZID == zID {
				f.deleteShare(token, false)
			}
		}
		delete(f.envs, zID)
		w.WriteHeader(http.StatusOK)
	case "POST /fake/enable":
		f.next++
		zID := fmt.Sprintf("env%d", f.next)
		f.envs[zID] = &fakeEnv{zID: zID, description: field("description")}
		writeJSON(w, http.StatusOK, map[string]string{"zId": zID})
	case "POST /fake/name":
		key := field("namespace") + ":" + field("name")
		if _, found := f.names[key]; found {
			w.WriteHeader(http.StatusConflict)
			return
		}
		f.names[key] = &fakeName{namespace: field("namespace"), name: field("name")}
		w.WriteHeader(http.StatusOK)
	case "POST /fake/share":
		name, found := f.names[field("name")]
		if !found || name.share != "" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		token := field("token")
		if _, used := f.shares[token]; token == "" || used {
			f.next++
			token = fmt.Sprintf("shr%d", f.next)
		}
		f.shares[token] = &fakeShare{token: token, envZID: field("envZId"), target: field("target"), name: field("name")}
		name.share = token
		writeJSON(w, http.StatusOK, map[string]string{"token": token, "endpoint": f.frontend.URL + "/f/" + token})
	case "POST /fake/delete":
		if _, found := f.shares[field("token")]; found {
			f.deleteShare(field("token"), field("keepBinding") == "true")
		}
		f.failReads += f.failOnStop
		w.WriteHeader(http.StatusOK)
	case "POST /fake/drop-name":
		delete(f.names, field("name"))
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *fakeController) deleteShare(token string, keepBinding bool) {
	shr := f.shares[token]
	delete(f.shares, token)
	if name, found := f.names[shr.name]; found && name.share == token && !keepBinding {
		name.share = ""
	}
}

func (f *fakeController) overview() *rest_model_zrok.Overview {
	ov := &rest_model_zrok.Overview{}
	for _, env := range f.envs {
		resources := &rest_model_zrok.EnvironmentAndResources{Environment: &rest_model_zrok.Environment{ZID: env.zID, Description: env.description}}
		for _, shr := range f.shares {
			if shr.envZID == env.zID {
				resources.Shares = append(resources.Shares, &rest_model_zrok.Share{ShareToken: shr.token, EnvZID: shr.envZID, ShareMode: "public"})
			}
		}
		ov.Environments = append(ov.Environments, resources)
	}
	return ov
}

func (f *fakeController) nameList() []*rest_model_zrok.Name {
	var out []*rest_model_zrok.Name
	for _, name := range f.names {
		out = append(out, &rest_model_zrok.Name{NamespaceToken: name.namespace, Name: name.name, Reserved: true, ShareToken: name.share})
	}
	return out
}

func (f *fakeController) serveFrontend(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/f/")
	token, path, _ := strings.Cut(rest, "/")
	f.mu.Lock()
	shr, found := f.shares[token]
	wrong := f.wrongBody
	f.mu.Unlock()
	if !found {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if wrong {
		_, _ = io.WriteString(w, "not the expected body")
		return
	}
	resp, err := http.Get(shr.target + "/" + path)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// fakeCLI stands in for the zrok2 binary in a subprocess: the harness launches it exactly as it launches the
// real one, and it changes the fake controller only through http.
func fakeCLI(args []string) int {
	modes := strings.Split(os.Getenv(fakeModeEnv), ",")
	api := os.Getenv("ZROK2_API_ENDPOINT")
	switch {
	case len(args) == 5 && args[0] == "enable":
		description, token := args[3], args[4]
		fmt.Printf("enabling with account token '%v'\n", token)
		var resp fakeEnableResponse
		if err := fakePost(api, token, "/fake/enable", map[string]string{"description": description}, &resp); err != nil {
			return fakeFail(err)
		}
		if slices.Contains(modes, "enable-lost") {
			return fakeFail(fmt.Errorf("lost the enable response"))
		}
		root, err := environment.LoadRoot()
		if err != nil {
			return fakeFail(err)
		}
		if err := root.SetEnvironment(&env_core.Environment{AccountToken: token, ZitiIdentity: resp.ZID, ApiEndpoint: api}); err != nil {
			return fakeFail(err)
		}
		return 0
	case len(args) == 5 && args[0] == "create" && args[1] == "name":
		root, err := environment.LoadRoot()
		if err != nil {
			return fakeFail(err)
		}
		if err := fakePost(api, root.Environment().AccountToken, "/fake/name", map[string]string{"namespace": args[3], "name": args[4]}, nil); err != nil {
			return fakeFail(err)
		}
		return 0
	case len(args) == 2 && args[0] == "agent" && args[1] == "start":
		return fakeAgent(api, modes)
	case len(args) == 7 && args[0] == "share" && args[1] == "public":
		root, err := environment.LoadRoot()
		if err != nil {
			return fakeFail(err)
		}
		client, conn, err := agentClient.NewClient(root)
		if err != nil {
			return fakeFail(err)
		}
		defer func() { _ = conn.Close() }()
		namespace, name, _ := strings.Cut(args[5], ":")
		resp, err := client.SharePublic(context.Background(), &agentGrpc.SharePublicRequest{
			Target:         args[6],
			NameSelections: []*agentGrpc.NameSelection{{NamespaceToken: namespace, Name: name}},
		})
		if err != nil {
			return fakeFail(err)
		}
		fmt.Println(resp)
		return 0
	}
	return fakeFail(fmt.Errorf("unexpected fake invocation with %d arguments", len(args)))
}

func fakeFail(err error) int {
	fmt.Fprintf(os.Stderr, "fake zrok2: %v\n", err)
	return 1
}

func fakePost(api, token, path string, body map[string]string, out any) error {
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, api+"/api/v2"+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("X-TOKEN", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("'%v' answered %d", path, resp.StatusCode)
	}
	if out != nil {
		var values map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
			return err
		}
		return dd.Bind(out, values)
	}
	return nil
}

type fakeEnableResponse struct {
	ZID string `dd:"zId"`
}

type fakeShareResponse struct {
	Token    string `dd:"token"`
	Endpoint string `dd:"endpoint"`
}

type fakeRegistry struct {
	Entries []fakeRegistryEntry `dd:"entries"`
}

type fakeRegistryEntry struct {
	Name   string `dd:"name"`
	Target string `dd:"target"`
	Token  string `dd:"token"`
}

// fakeAgentServer keeps a registry next to the environment, replays it at start and deletes its shares on
// interrupt, with modes that break each of those in the ways the harness must detect.
type fakeAgentServer struct {
	agentGrpc.UnimplementedAgentServer
	mu       sync.Mutex
	api      string
	root     env_core.Root
	modes    []string
	registry string
	entries  []fakeRegistryEntry
	shares   map[string]*agentGrpc.ShareDetail
}

func fakeAgent(api string, modes []string) int {
	// hang swallows the interrupt rather than ignoring it: on windows an ignored console event falls through to
	// the default handler, which exits the process
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	root, err := environment.LoadRoot()
	if err != nil {
		return fakeFail(err)
	}
	socket, err := root.AgentSocket()
	if err != nil {
		return fakeFail(err)
	}
	_ = os.Remove(socket)
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return fakeFail(err)
	}
	a := &fakeAgentServer{api: api, root: root, modes: modes, registry: filepath.Join(filepath.Dir(socket), "fake-registry.json"), shares: make(map[string]*agentGrpc.ShareDetail)}
	var saved fakeRegistry
	if err := dd.BindFromJSON(&saved, a.registry); err == nil {
		a.entries = saved.Entries
	}
	for i := range a.entries {
		token := ""
		if slices.Contains(modes, "same-token") {
			token = a.entries[i].Token
		}
		if err := a.create(&a.entries[i], token); err != nil {
			return fakeFail(err)
		}
	}
	a.save()
	server := grpc.NewServer()
	agentGrpc.RegisterAgentServer(server, a)
	go func() { _ = server.Serve(listener) }()
	<-interrupts
	if slices.Contains(modes, "hang") {
		select {}
	}

	a.mu.Lock()
	for token := range a.shares {
		if slices.Contains(modes, "leak-token") {
			continue
		}
		keep := fmt.Sprint(slices.Contains(modes, "keep-binding"))
		_ = fakePost(api, root.Environment().AccountToken, "/fake/delete", map[string]string{"token": token, "keepBinding": keep}, nil)
	}
	if slices.Contains(modes, "drop-name") {
		for _, entry := range a.entries {
			_ = fakePost(api, root.Environment().AccountToken, "/fake/drop-name", map[string]string{"name": entry.Name}, nil)
		}
	}
	a.mu.Unlock()
	server.Stop()
	if slices.Contains(modes, "exit-nonzero") {
		return 1
	}
	return 0
}

func (a *fakeAgentServer) create(entry *fakeRegistryEntry, token string) error {
	var resp fakeShareResponse
	body := map[string]string{"envZId": a.root.Environment().ZitiIdentity, "name": entry.Name, "target": entry.Target, "token": token}
	if err := fakePost(a.api, a.root.Environment().AccountToken, "/fake/share", body, &resp); err != nil {
		return err
	}
	entry.Token = resp.Token
	a.shares[resp.Token] = &agentGrpc.ShareDetail{Token: resp.Token, ShareMode: "public", BackendMode: "proxy", BackendEndpoint: entry.Target, FrontendEndpoint: []string{resp.Endpoint}, Status: "active"}
	return nil
}

func (a *fakeAgentServer) save() {
	data, _ := marshalJSON(&fakeRegistry{Entries: a.entries})
	_ = os.WriteFile(a.registry, data, 0o600)
}

func (a *fakeAgentServer) Version(context.Context, *agentGrpc.VersionRequest) (*agentGrpc.VersionResponse, error) {
	return &agentGrpc.VersionResponse{V: os.Getenv(fakeBuildEnv)}, nil
}

func (a *fakeAgentServer) Status(context.Context, *agentGrpc.StatusRequest) (*agentGrpc.StatusResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	resp := &agentGrpc.StatusResponse{}
	for _, shr := range a.shares {
		resp.Shares = append(resp.Shares, shr)
	}
	return resp, nil
}

func (a *fakeAgentServer) SharePublic(_ context.Context, req *agentGrpc.SharePublicRequest) (*agentGrpc.SharePublicResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	selection := req.GetNameSelections()[0]
	entry := fakeRegistryEntry{Name: selection.GetNamespaceToken() + ":" + selection.GetName(), Target: req.GetTarget()}
	if err := a.create(&entry, ""); err != nil {
		return nil, err
	}
	a.entries = append(a.entries, entry)
	a.save()
	return &agentGrpc.SharePublicResponse{Token: entry.Token, FrontendEndpoints: a.shares[entry.Token].FrontendEndpoint}, nil
}
