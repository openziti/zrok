package private_frontend

type Config struct {
	IdentityName string
	SvcName      string
	Address      string
}

func DefaultConfig(identityName string) *Config {
	return &Config{
		IdentityName: identityName,
		Address:      "0.0.0.0:8080",
	}
}
