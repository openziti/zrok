package dynamicProxy

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type authHandler struct {
	cfg        *config
	signingKey []byte
	handler    http.Handler
}

func newAuthHandler(cfg *config, signingKey []byte, handler http.Handler) *authHandler {
	return &authHandler{
		cfg:        cfg,
		signingKey: signingKey,
		handler:    handler,
	}
}

type IntermediateJWT struct {
	State           string `json:"st"`
	TargetHost      string `json:"th"`
	RefreshInterval string `json:"rfi"`
	jwt.RegisteredClaims
}
