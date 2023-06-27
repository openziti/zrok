package emailUi

type Config struct {
	Host     string
	Port     int
	Username string
	Password string `cf:"+secret"`
	From     string
}
