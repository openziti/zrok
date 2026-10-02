package zitifake

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
)

// kinds of object the fake serves, named by their edge-management collection path.
const (
	Configs                   = "configs"
	Services                  = "services"
	ServicePolicies           = "service-policies"
	ServiceEdgeRouterPolicies = "service-edge-router-policies"
	Identities                = "identities"
	EdgeRouterPolicies        = "edge-router-policies"
)

// ziti's page sizes: the page served for a missing or zero limit, and the largest it serves.
const (
	defaultPageSize = 10
	maxPageSize     = 500
)

var labels = map[string]struct{ id, noun string }{
	Configs:                   {"config", "config"},
	Services:                  {"service", "service"},
	ServicePolicies:           {"policy", "service policy"},
	ServiceEdgeRouterPolicies: {"serp", "service edge router policy"},
	Identities:                {"identity", "identity"},
	EdgeRouterPolicies:        {"erp", "edge router policy"},
}

type object struct {
	name   string
	tags   *rest_model.Tags
	dial   *rest_model.DialBind
	detail interface{}
}

// Server implements the edge-management resources used by controller tests.
type Server struct {
	*httptest.Server
	mu                             sync.Mutex
	objects                        map[string]map[string]*object
	nextID                         int
	PolicyCreates, PolicyDeletes   int
	ServiceCreates, ServiceDeletes int
	ConfigCreates, ConfigDeletes   int
	SerpCreates, SerpDeletes       int
	createLog, deleteLog           []string
	beforeCreate                   func(kind, name string)
	username, password             string
	sessions                       map[string]bool
	nextToken                      int
	authentications, authAttempts  int
	unauthorized                   int
	rejectAuthentication           bool
	rejectOperations               bool
	rejectDeletes                  map[string]bool
	authDelay                      time.Duration
}

func New() *Server {
	return NewWithCredentials("admin", "admin")
}

func NewWithCredentials(username, password string) *Server {
	f := newServer(username, password)
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	return f
}

// NewTLSWithCredentials serves over https with the httptest certificate, which Certificate returns.
func NewTLSWithCredentials(username, password string) *Server {
	f := newServer(username, password)
	f.Server = httptest.NewTLSServer(http.HandlerFunc(f.serve))
	return f
}

func newServer(username, password string) *Server {
	objects := make(map[string]map[string]*object)
	for _, kind := range []string{Configs, Services, ServicePolicies, ServiceEdgeRouterPolicies, Identities, EdgeRouterPolicies} {
		objects[kind] = make(map[string]*object)
	}
	return &Server{objects: objects, sessions: make(map[string]bool), rejectDeletes: make(map[string]bool), username: username, password: password}
}

func (f *Server) Edge() *rest_management_api_client.ZitiEdgeManagement {
	// resource-only tests receive an already authenticated fixture client.
	f.mu.Lock()
	token := f.issueToken()
	f.mu.Unlock()
	scheme, host, _ := strings.Cut(f.URL, "://")
	runtime := httptransport.NewWithClient(host, "/edge/management/v1", []string{scheme}, f.Client())
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

// RejectDeletes answers every delete of kind with an internal server error.
func (f *Server) RejectDeletes(kind string, reject bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejectDeletes[kind] = reject
}

// SetAuthenticationDelay delays every authentication response without blocking other requests.
func (f *Server) SetAuthenticationDelay(delay time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.authDelay = delay
}

func (f *Server) AuthCounts() (successful, attempts, unauthorized int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.authentications, f.authAttempts, f.unauthorized
}

// OnBeforeCreate installs a hook invoked with the kind and name of each create before it is applied.
// the hook runs under the fake's lock, so it may call Seed but no other method.
func (f *Server) OnBeforeCreate(hook func(kind, name string)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.beforeCreate = hook
}

// Seed inserts an object directly, without counting it as a create. it expects the fake's lock to be
// held, so it is called from a BeforeCreate hook.
func (f *Server) Seed(kind, name string, tags *rest_model.Tags) string {
	f.nextID++
	id := fmt.Sprintf("%s-%d", labels[kind].id, f.nextID)
	f.seed(kind, id, name, tags)
	return id
}

// SeedWithID inserts an object under a caller-chosen id, without counting it as a create. it takes the
// fake's lock, so it is called from a test rather than a BeforeCreate hook. identities are addressed
// by the id the store records (an environment's ZId), which is why this form exists.
func (f *Server) SeedWithID(kind, id, name string, tags *rest_model.Tags) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seed(kind, id, name, tags)
}

// Backdate moves the createdAt of the object of kind with id age into the past.
func (f *Server) Backdate(kind, id string, age time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	createdAt := strfmt.DateTime(time.Now().Add(-age))
	baseOf(f.objects[kind][id].detail).CreatedAt = &createdAt
}

// Len reports how many objects of kind the fake holds.
func (f *Server) Len(kind string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.objects[kind])
}

// Has reports whether an object of kind exists with id.
func (f *Server) Has(kind, id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.objects[kind][id] != nil
}

func (f *Server) seed(kind, id, name string, tags *rest_model.Tags) {
	obj := &object{name: name, tags: tags}
	switch kind {
	case Configs:
		obj.detail = &rest_model.ConfigDetail{BaseEntity: base(id, tags), Name: &obj.name, ConfigType: &rest_model.EntityRef{}, ConfigTypeID: new(string), Data: map[string]interface{}{}}
	case Services:
		obj.detail = &rest_model.ServiceDetail{BaseEntity: base(id, tags), Name: &obj.name, Config: map[string]map[string]interface{}{}, Configs: []string{}, EncryptionRequired: new(bool), MaxIdleTimeMillis: new(int64), Permissions: rest_model.DialBindArray{}, PostureQueries: []*rest_model.PostureQueries{}, RoleAttributes: &rest_model.Attributes{}, TerminatorStrategy: new(string)}
	case ServicePolicies:
		dial := rest_model.DialBindBind
		obj.dial = &dial
		obj.detail = &rest_model.ServicePolicyDetail{BaseEntity: base(id, tags), Name: &obj.name, IdentityRoles: rest_model.Roles{}, IdentityRolesDisplay: rest_model.NamedRoles{}, PostureCheckRoles: rest_model.Roles{}, PostureCheckRolesDisplay: rest_model.NamedRoles{}, ServiceRoles: rest_model.Roles{}, ServiceRolesDisplay: rest_model.NamedRoles{}, Semantic: new(rest_model.SemanticAllOf), Type: obj.dial}
	case ServiceEdgeRouterPolicies:
		obj.detail = &rest_model.ServiceEdgeRouterPolicyDetail{BaseEntity: base(id, tags), Name: &obj.name, EdgeRouterRoles: rest_model.Roles{}, EdgeRouterRolesDisplay: rest_model.NamedRoles{}, ServiceRoles: rest_model.Roles{}, ServiceRolesDisplay: rest_model.NamedRoles{}, Semantic: new(rest_model.SemanticAllOf)}
	case Identities:
		obj.detail = &rest_model.IdentityDetail{BaseEntity: base(id, tags), Name: &obj.name, AuthPolicy: &rest_model.EntityRef{}, AuthPolicyID: new(string), Authenticators: &rest_model.IdentityAuthenticators{}, DefaultHostingCost: new(rest_model.TerminatorCost), Disabled: new(bool), EdgeRouterConnectionStatus: new(string), Enrollment: &rest_model.IdentityEnrollments{}, EnvInfo: &rest_model.EnvInfo{}, ExternalID: new(string), HasAPISession: new(bool), HasEdgeRouterConnection: new(bool), Interfaces: []*rest_model.Interface{}, IsAdmin: new(bool), IsDefaultAdmin: new(bool), IsMfaEnabled: new(bool), RoleAttributes: &rest_model.Attributes{}, SdkInfo: &rest_model.SdkInfo{}, ServiceHostingCosts: rest_model.TerminatorCostMap{}, ServiceHostingPrecedences: rest_model.TerminatorPrecedenceMap{}, Type: &rest_model.EntityRef{}, TypeID: new(string)}
	case EdgeRouterPolicies:
		obj.detail = &rest_model.EdgeRouterPolicyDetail{BaseEntity: base(id, tags), Name: &obj.name, EdgeRouterRoles: rest_model.Roles{}, EdgeRouterRolesDisplay: rest_model.NamedRoles{}, IdentityRoles: rest_model.Roles{}, IdentityRolesDisplay: rest_model.NamedRoles{}, IsSystem: new(bool), Semantic: new(rest_model.SemanticAllOf)}
	default:
		panic("unknown kind " + kind)
	}
	f.objects[kind][id] = obj
}

// Log returns the objects created and deleted through the api, in order, as 'kind/id'.
func (f *Server) Log() (created, deleted []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.createLog...), append([]string(nil), f.deleteLog...)
}

// Tagged returns 'kind/id' for every object carrying tag key=value, sorted.
func (f *Server) Tagged(key, value string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for kind, objects := range f.objects {
		for id, obj := range objects {
			if obj.tags != nil && fmt.Sprint(obj.tags.SubTags[key]) == value {
				out = append(out, kind+"/"+id)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Remove deletes an object without counting or logging the delete.
func (f *Server) Remove(kind, id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.objects[kind], id)
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
	path := strings.TrimPrefix(r.URL.Path, "/edge/management/v1/")
	if path == "authenticate" {
		f.mu.Lock()
		delay := f.authDelay
		f.mu.Unlock()
		time.Sleep(delay)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
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
	kind := parts[0]
	if f.objects[kind] == nil {
		writeError(w, 404, "route not found")
		return
	}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			f.list(w, r, kind)
		case http.MethodPost:
			if kind == Identities || kind == EdgeRouterPolicies {
				writeError(w, 405, "method not allowed")
				return
			}
			f.create(w, r, kind)
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
		f.detail(w, kind, id)
	case http.MethodDelete:
		f.delete(w, kind, id)
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

func baseOf(detail interface{}) *rest_model.BaseEntity {
	switch v := detail.(type) {
	case *rest_model.ConfigDetail:
		return &v.BaseEntity
	case *rest_model.ServiceDetail:
		return &v.BaseEntity
	case *rest_model.ServicePolicyDetail:
		return &v.BaseEntity
	case *rest_model.ServiceEdgeRouterPolicyDetail:
		return &v.BaseEntity
	case *rest_model.IdentityDetail:
		return &v.BaseEntity
	case *rest_model.EdgeRouterPolicyDetail:
		return &v.BaseEntity
	default:
		panic(fmt.Sprintf("unknown detail %T", detail))
	}
}

func (f *Server) create(w http.ResponseWriter, r *http.Request, kind string) {
	var name string
	var tags *rest_model.Tags
	var fill func(obj *object, id string)
	switch kind {
	case Configs:
		var input rest_model.ConfigCreate
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == nil || input.ConfigTypeID == nil {
			writeError(w, 400, "invalid config")
			return
		}
		name, tags = *input.Name, input.Tags
		fill = func(obj *object, id string) {
			obj.detail = &rest_model.ConfigDetail{BaseEntity: base(id, tags), Name: &obj.name, ConfigType: &rest_model.EntityRef{ID: *input.ConfigTypeID}, ConfigTypeID: input.ConfigTypeID, Data: input.Data}
		}
	case Services:
		var input rest_model.ServiceCreate
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == nil {
			writeError(w, 400, "invalid service")
			return
		}
		name, tags = *input.Name, input.Tags
		fill = func(obj *object, id string) {
			maxIdle := input.MaxIdleTimeMillis
			strategy := input.TerminatorStrategy
			roles := rest_model.Attributes(input.RoleAttributes)
			obj.detail = &rest_model.ServiceDetail{BaseEntity: base(id, tags), Name: &obj.name, Config: map[string]map[string]interface{}{}, Configs: input.Configs, EncryptionRequired: input.EncryptionRequired, MaxIdleTimeMillis: &maxIdle, Permissions: rest_model.DialBindArray{}, PostureQueries: []*rest_model.PostureQueries{}, RoleAttributes: &roles, TerminatorStrategy: &strategy}
		}
	case ServicePolicies:
		var input rest_model.ServicePolicyCreate
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == nil {
			writeError(w, 400, "invalid service policy")
			return
		}
		name, tags = *input.Name, input.Tags
		fill = func(obj *object, id string) {
			obj.dial = input.Type
			obj.detail = &rest_model.ServicePolicyDetail{BaseEntity: base(id, tags), Name: &obj.name, IdentityRoles: input.IdentityRoles, IdentityRolesDisplay: rest_model.NamedRoles{}, PostureCheckRoles: input.PostureCheckRoles, PostureCheckRolesDisplay: rest_model.NamedRoles{}, ServiceRoles: input.ServiceRoles, ServiceRolesDisplay: rest_model.NamedRoles{}, Semantic: input.Semantic, Type: input.Type}
		}
	case ServiceEdgeRouterPolicies:
		var input rest_model.ServiceEdgeRouterPolicyCreate
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == nil {
			writeError(w, 400, "invalid service edge router policy")
			return
		}
		name, tags = *input.Name, input.Tags
		fill = func(obj *object, id string) {
			obj.detail = &rest_model.ServiceEdgeRouterPolicyDetail{BaseEntity: base(id, tags), Name: &obj.name, EdgeRouterRoles: input.EdgeRouterRoles, EdgeRouterRolesDisplay: rest_model.NamedRoles{}, ServiceRoles: input.ServiceRoles, ServiceRolesDisplay: rest_model.NamedRoles{}, Semantic: input.Semantic}
		}
	}
	if f.beforeCreate != nil {
		f.beforeCreate(kind, name)
	}
	for _, existing := range f.objects[kind] {
		if existing.name == name {
			writeError(w, 400, labels[kind].noun+" name conflict: "+name)
			return
		}
	}
	id := f.Seed(kind, name, tags)
	fill(f.objects[kind][id], id)
	f.createLog = append(f.createLog, kind+"/"+id)
	switch kind {
	case Configs:
		f.ConfigCreates++
	case Services:
		f.ServiceCreates++
	case ServicePolicies:
		f.PolicyCreates++
	case ServiceEdgeRouterPolicies:
		f.SerpCreates++
	}
	writeJSON(w, 201, &rest_model.CreateEnvelope{Data: &rest_model.CreateLocation{ID: id}, Meta: &rest_model.Meta{}})
}

// list pages like ziti: a missing or zero limit is the default page of ten, a larger one is capped at
// 500, and offset skips into the matches in id order.
func (f *Server) list(w http.ResponseWriter, r *http.Request, kind string) {
	query := r.URL.Query()
	filter := query.Get("filter")
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = defaultPageSize
	}
	limit = min(limit, maxPageSize)
	offset, _ := strconv.Atoi(query.Get("offset"))
	ids := make([]string, 0, len(f.objects[kind]))
	for id := range f.objects[kind] {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var matched []*object
	for _, id := range ids {
		obj := f.objects[kind][id]
		if match(filter, id, obj.name, obj.tags, obj.dial) {
			matched = append(matched, obj)
		}
	}
	matched = matched[min(offset, len(matched)):]
	matched = matched[:min(limit, len(matched))]
	switch kind {
	case Configs:
		writeJSON(w, 200, &rest_model.ListConfigsEnvelope{Data: details[rest_model.ConfigDetail](matched), Meta: &rest_model.Meta{}})
	case Services:
		writeJSON(w, 200, &rest_model.ListServicesEnvelope{Data: details[rest_model.ServiceDetail](matched), Meta: &rest_model.Meta{}})
	case ServicePolicies:
		writeJSON(w, 200, &rest_model.ListServicePoliciesEnvelope{Data: details[rest_model.ServicePolicyDetail](matched), Meta: &rest_model.Meta{}})
	case ServiceEdgeRouterPolicies:
		writeJSON(w, 200, &rest_model.ListServiceEdgeRouterPoliciesEnvelope{Data: details[rest_model.ServiceEdgeRouterPolicyDetail](matched), Meta: &rest_model.Meta{}})
	case Identities:
		writeJSON(w, 200, &rest_model.ListIdentitiesEnvelope{Data: details[rest_model.IdentityDetail](matched), Meta: &rest_model.Meta{}})
	case EdgeRouterPolicies:
		writeJSON(w, 200, &rest_model.ListEdgeRouterPoliciesEnvelope{Data: details[rest_model.EdgeRouterPolicyDetail](matched), Meta: &rest_model.Meta{}})
	}
}

func details[T any](objects []*object) []*T {
	out := make([]*T, 0, len(objects))
	for _, obj := range objects {
		out = append(out, obj.detail.(*T))
	}
	return out
}

func (f *Server) detail(w http.ResponseWriter, kind, id string) {
	obj := f.objects[kind][id]
	if obj == nil {
		writeError(w, 404, "resource not found: "+id)
		return
	}
	switch kind {
	case Configs:
		writeJSON(w, 200, &rest_model.DetailConfigEnvelope{Data: obj.detail.(*rest_model.ConfigDetail), Meta: &rest_model.Meta{}})
	case Services:
		writeJSON(w, 200, &rest_model.DetailServiceEnvelope{Data: obj.detail.(*rest_model.ServiceDetail), Meta: &rest_model.Meta{}})
	case ServicePolicies:
		writeJSON(w, 200, &rest_model.DetailServicePolicyEnvelop{Data: obj.detail.(*rest_model.ServicePolicyDetail), Meta: &rest_model.Meta{}})
	case ServiceEdgeRouterPolicies:
		writeJSON(w, 200, &rest_model.DetailServiceEdgePolicyEnvelope{Data: obj.detail.(*rest_model.ServiceEdgeRouterPolicyDetail), Meta: &rest_model.Meta{}})
	case Identities:
		writeJSON(w, 200, &rest_model.DetailIdentityEnvelope{Data: obj.detail.(*rest_model.IdentityDetail), Meta: &rest_model.Meta{}})
	case EdgeRouterPolicies:
		writeJSON(w, 200, &rest_model.DetailEdgeRouterPolicyEnvelope{Data: obj.detail.(*rest_model.EdgeRouterPolicyDetail), Meta: &rest_model.Meta{}})
	}
}

func (f *Server) delete(w http.ResponseWriter, kind, id string) {
	if f.rejectDeletes[kind] {
		writeError(w, http.StatusInternalServerError, labels[kind].noun+" delete rejected: "+id)
		return
	}
	if f.objects[kind][id] == nil {
		writeError(w, 404, labels[kind].noun+" not found: "+id)
		return
	}
	delete(f.objects[kind], id)
	f.deleteLog = append(f.deleteLog, kind+"/"+id)
	switch kind {
	case Configs:
		f.ConfigDeletes++
	case Services:
		f.ServiceDeletes++
	case ServicePolicies:
		f.PolicyDeletes++
	case ServiceEdgeRouterPolicies:
		f.SerpDeletes++
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
