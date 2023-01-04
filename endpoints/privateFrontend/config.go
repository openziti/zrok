package privateFrontend

type Config struct {
	IdentityName string
	ShrToken     string
	Address      string
}

func DefaultConfig(identityName string) *Config {
	return &Config{
		IdentityName: identityName,
		Address:      "0.0.0.0:8080",
	}
}
