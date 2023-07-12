package env_core

type Environment struct {
	Token        string
	ZitiIdentity string
	ApiEndpoint  string
}

type Config struct {
	ApiEndpoint string
}

type Metadata struct {
	V        string
	RootPath string
}
