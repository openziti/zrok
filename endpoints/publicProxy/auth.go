package publicProxy

import "github.com/golang-jwt/jwt/v5"

type authHandler struct {
	cfg *Config
	key []byte
}

func newAuthHandler(cfg *Config, key []byte) *authHandler {
	return &authHandler{
		cfg: cfg,
		key: key,
	}
}

type IntermediateJWT struct {
	State                      string `json:"state"`
	Host                       string `json:"host"`
	AuthorizationCheckInterval string `json:"authorizationCheckInterval"`
	jwt.RegisteredClaims
}
