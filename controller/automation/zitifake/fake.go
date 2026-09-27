package zitifake

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
)

// Server implements the edge-management resources used by controller tests.
type Server struct {
	*httptest.Server
	mu                             sync.Mutex
	policies                       map[string]*rest_model.ServicePolicyDetail
	services                       map[string]*rest_model.ServiceDetail
	nextID                         int
	PolicyCreates, PolicyDeletes   int
	ServiceCreates, ServiceDeletes int
	username, password             string
	sessions                       map[string]bool
	nextToken                      int
	authentications, authAttempts  int
	unauthorized                   int
	rejectAuthentication           bool
	rejectOperations               bool
}

func New() *Server {
	return NewWithCredentials("admin", "admin")
}

func NewWithCredentials(username, password string) *Server {
	f := &Server{policies: make(map[string]*rest_model.ServicePolicyDetail), services: make(map[string]*rest_model.ServiceDetail), sessions: make(map[string]bool), username: username, password: password}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	return f
}

func (f *Server) Edge() *rest_management_api_client.ZitiEdgeManagement {
	// resource-only tests receive an already authenticated fixture client.
	f.mu.Lock()
	token := f.issueToken()
	f.mu.Unlock()
	host := strings.TrimPrefix(f.URL, "http://")
	runtime := httptransport.New(host, "/edge/management/v1", []string{"http"})
	runtime.DefaultAuthentication = &rest_util.ZitiTokenAuth{Token: token}
	return rest_management_api_client.New(runtime, nil)
}

func (f *Server) ExpireAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	clear(f.sessions)
}

func (f *Server) RejectAuthentication(reject bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejectAuthentication = reject
}

func (f *Server) RejectOperations(reject bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejectOperations = reject
}

func (f *Server) AuthCounts() (successful, attempts, unauthorized int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.authentications, f.authAttempts, f.unauthorized
}

func (f *Server) issueToken() string {
	f.nextToken++
	token := fmt.Sprintf("session-%d", f.nextToken)
	f.sessions[token] = true
	return token
}

func (f *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	f.authAttempts++
	var input rest_model.Authenticate
	if r.Method != http.MethodPost || r.URL.Query().Get("method") != "password" || json.NewDecoder(r.Body).Decode(&input) != nil || string(input.Username) != f.username || string(input.Password) != f.password || f.rejectAuthentication {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token := f.issueToken()
	f.authentications++
	writeJSON(w, http.StatusOK, &rest_model.CurrentAPISessionDetailEnvelope{
		Data: &rest_model.CurrentAPISessionDetail{APISessionDetail: rest_model.APISessionDetail{Token: &token}}, Meta: &rest_model.Meta{},
	})
}

func (f *Server) Counts() (policyCreates, policyDeletes, serviceCreates, serviceDeletes int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.PolicyCreates, f.PolicyDeletes, f.ServiceCreates, f.ServiceDeletes
}

func (f *Server) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/edge/management/v1/")
	if path == "authenticate" {
		f.authenticate(w, r)
		return
	}
	if !f.sessions[r.Header.Get("zt-session")] || f.rejectOperations {
		f.unauthorized++
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	parts := strings.Split(path, "/")
	if len(parts) < 1 || (parts[0] != "service-policies" && parts[0] != "services") {
		writeError(w, 404, "route not found")
		return
	}
	policy := parts[0] == "service-policies"
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			f.list(w, r, policy)
		case http.MethodPost:
			f.create(w, r, policy)
		default:
			writeError(w, 405, "method not allowed")
		}
		return
	}
	if len(parts) != 2 {
		writeError(w, 404, "route not found")
		return
	}
	id := parts[1]
	switch r.Method {
	case http.MethodGet:
		f.detail(w, policy, id)
	case http.MethodDelete:
		f.delete(w, policy, id)
	default:
		writeError(w, 405, "method not allowed")
	}
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, &rest_model.APIErrorEnvelope{Error: &rest_model.APIError{Code: http.StatusText(status), Message: message}, Meta: &rest_model.Meta{}})
}

func base(id string, tags *rest_model.Tags) rest_model.BaseEntity {
	now := strfmt.DateTime(time.Now())
	return rest_model.BaseEntity{ID: &id, Tags: tags, CreatedAt: &now, UpdatedAt: &now, Links: rest_model.Links{}}
}

func (f *Server) create(w http.ResponseWriter, r *http.Request, policy bool) {
	var name string
	var tags *rest_model.Tags
	if policy {
		var input rest_model.ServicePolicyCreate
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == nil {
			writeError(w, 400, "invalid service policy")
			return
		}
		name, tags = *input.Name, input.Tags
		for _, existing := range f.policies {
			if *existing.Name == name {
				writeError(w, 400, "service policy name conflict: "+name)
				return
			}
		}
		f.nextID++
		id := fmt.Sprintf("policy-%d", f.nextID)
		f.policies[id] = &rest_model.ServicePolicyDetail{BaseEntity: base(id, tags), Name: &name, IdentityRoles: input.IdentityRoles, IdentityRolesDisplay: rest_model.NamedRoles{}, PostureCheckRoles: input.PostureCheckRoles, PostureCheckRolesDisplay: rest_model.NamedRoles{}, ServiceRoles: input.ServiceRoles, ServiceRolesDisplay: rest_model.NamedRoles{}, Semantic: input.Semantic, Type: input.Type}
		f.PolicyCreates++
		writeJSON(w, 201, &rest_model.CreateEnvelope{Data: &rest_model.CreateLocation{ID: id}, Meta: &rest_model.Meta{}})
		return
	}
	var input rest_model.ServiceCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == nil {
		writeError(w, 400, "invalid service")
		return
	}
	name, tags = *input.Name, input.Tags
	for _, existing := range f.services {
		if *existing.Name == name {
			writeError(w, 400, "service name conflict: "+name)
			return
		}
	}
	f.nextID++
	id := fmt.Sprintf("service-%d", f.nextID)
	maxIdle := input.MaxIdleTimeMillis
	strategy := input.TerminatorStrategy
	roles := rest_model.Attributes(input.RoleAttributes)
	f.services[id] = &rest_model.ServiceDetail{BaseEntity: base(id, tags), Name: &name, Config: map[string]map[string]interface{}{}, Configs: input.Configs, EncryptionRequired: input.EncryptionRequired, MaxIdleTimeMillis: &maxIdle, Permissions: rest_model.DialBindArray{}, PostureQueries: []*rest_model.PostureQueries{}, RoleAttributes: &roles, TerminatorStrategy: &strategy}
	f.ServiceCreates++
	writeJSON(w, 201, &rest_model.CreateEnvelope{Data: &rest_model.CreateLocation{ID: id}, Meta: &rest_model.Meta{}})
}

func (f *Server) list(w http.ResponseWriter, r *http.Request, policy bool) {
	filter := r.URL.Query().Get("filter")
	if policy {
		ids := make([]string, 0, len(f.policies))
		for id := range f.policies {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		data := rest_model.ServicePolicyList{}
		for _, id := range ids {
			p := f.policies[id]
			if match(filter, id, *p.Name, p.Tags, p.Type) {
				data = append(data, p)
			}
		}
		writeJSON(w, 200, &rest_model.ListServicePoliciesEnvelope{Data: data, Meta: &rest_model.Meta{}})
		return
	}
	ids := make([]string, 0, len(f.services))
	for id := range f.services {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	data := rest_model.ServiceList{}
	for _, id := range ids {
		s := f.services[id]
		if match(filter, id, *s.Name, s.Tags, nil) {
			data = append(data, s)
		}
	}
	writeJSON(w, 200, &rest_model.ListServicesEnvelope{Data: data, Meta: &rest_model.Meta{}})
}

func (f *Server) detail(w http.ResponseWriter, policy bool, id string) {
	if policy {
		if p := f.policies[id]; p != nil {
			writeJSON(w, 200, &rest_model.DetailServicePolicyEnvelop{Data: p, Meta: &rest_model.Meta{}})
			return
		}
	} else if s := f.services[id]; s != nil {
		writeJSON(w, 200, &rest_model.DetailServiceEnvelope{Data: s, Meta: &rest_model.Meta{}})
		return
	}
	writeError(w, 404, "resource not found: "+id)
}

func (f *Server) delete(w http.ResponseWriter, policy bool, id string) {
	if policy {
		if f.policies[id] == nil {
			writeError(w, 404, "service policy not found: "+id)
			return
		}
		delete(f.policies, id)
		f.PolicyDeletes++
	} else {
		if f.services[id] == nil {
			writeError(w, 404, "service not found: "+id)
			return
		}
		delete(f.services, id)
		f.ServiceDeletes++
	}
	writeJSON(w, 200, struct {
		Data interface{}      `json:"data"`
		Meta *rest_model.Meta `json:"meta"`
	}{Meta: &rest_model.Meta{}})
}

func match(filter, id, name string, tags *rest_model.Tags, kind *rest_model.DialBind) bool {
	if filter == "" {
		return true
	}
	for _, clause := range strings.Split(filter, " and ") {
		clause = strings.TrimSpace(clause)
		switch {
		case strings.HasPrefix(clause, "name="):
			if name != strings.Trim(clause[5:], `"`) {
				return false
			}
		case strings.HasPrefix(clause, "id="):
			if id != strings.Trim(clause[3:], `"`) {
				return false
			}
		case strings.HasPrefix(clause, "tags.") && strings.HasSuffix(clause, " != null"):
			key := strings.TrimSuffix(strings.TrimPrefix(clause, "tags."), " != null")
			if tags == nil || tags.SubTags[key] == nil {
				return false
			}
		case strings.HasPrefix(clause, "tags."):
			key, value, ok := strings.Cut(strings.TrimPrefix(clause, "tags."), "=")
			if !ok || tags == nil || fmt.Sprint(tags.SubTags[key]) != strings.Trim(value, `"`) {
				return false
			}
		case clause == "type=1":
			if kind == nil || *kind != rest_model.DialBindDial {
				return false
			}
		case clause == "type=2":
			if kind == nil || *kind != rest_model.DialBindBind {
				return false
			}
		default:
			return false
		}
	}
	return true
}
