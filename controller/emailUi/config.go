package emailUi

type Config struct {
	Host     string
	Port     int
	Username string
	Password string `dd:"+secret"`
	From     string
}
