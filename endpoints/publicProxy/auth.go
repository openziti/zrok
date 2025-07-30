package publicProxy

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

type oauthConfigurer interface {
	configure() error
}
