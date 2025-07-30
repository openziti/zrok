package publicProxy

import "github.com/golang-jwt/jwt/v5"

type authHandler struct {
	cfg        *Config
	signingKey []byte
}

func newAuthHandler(cfg *Config, signingKey []byte) *authHandler {
	return &authHandler{
		cfg:        cfg,
		signingKey: signingKey,
	}
}

type IntermediateJWT struct {
	State                      string `json:"state"`
	Host                       string `json:"host"`
	AuthorizationCheckInterval string `json:"authorizationCheckInterval"`
	jwt.RegisteredClaims
}
