package dynamicProxy

import "github.com/golang-jwt/jwt/v5"

type authHandler struct {
	cfg        *config
	signingKey []byte
}

func newAuthHandler(cfg *config, signingKey []byte) *authHandler {
	return &authHandler{
		cfg:        cfg,
		signingKey: signingKey,
	}
}

type IntermediateJWT struct {
	State           string `json:"st"`
	TargetHost      string `json:"th"`
	RefreshInterval string `json:"rfi"`
	ReturnToken     bool   `json:"rt,omitempty"`
	jwt.RegisteredClaims
}

// sessionRefresher is implemented by OIDC providers that support inline token refresh.
type sessionRefresher interface {
	RefreshSessionJWT(claims *zrokClaims) (string, error)
}
