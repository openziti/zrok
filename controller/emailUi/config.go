package emailUi

type Config struct {
	Host     string
	Port     int
	Username string
	Password string `df:",secret"`
	From     string
}
