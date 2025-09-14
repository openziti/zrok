package dynamicProxyController

type Config struct {
	IdentityPath  string               `dd:"+required"`
	ServiceName   string               `dd:"+required"`
	AmqpPublisher *AmqpPublisherConfig `dd:"+required"`
}
