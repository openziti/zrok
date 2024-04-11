package vpn

type ClientConfig struct {
	Greeting string
	IP       string
	CIDR     string
	ServerIP string
	Routes   []string
}
