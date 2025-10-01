package dynamicProxy

import "net/http"

func basicAuthRequired(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	_, _ = w.Write([]byte("No Authorization\n"))
}

func (h *authHandler) handleBasicAuth(w http.ResponseWriter, r *http.Request, cfg map[string]interface{}, shrToken string) bool {
	inUser, inPass, ok := r.BasicAuth()
	if !ok {
		basicAuthRequired(w, shrToken)
		return false
	}

	if v, found := cfg["basic_auth"]; found {
		if basicAuth, ok := v.(map[string]interface{}); ok {
			if users, found := basicAuth["users"].([]interface{}); found {
				for _, v := range users {
					if um, ok := v.(map[string]interface{}); ok {
						if um["username"] == inUser && um["password"] == inPass {
							r.Header.Set("zrok-auth-provider", "basic")
							r.Header.Set("zrok-auth-user", inUser)
							return true
						}
					}
				}
			}
		}
	}

	basicAuthRequired(w, shrToken)
	return false
}
