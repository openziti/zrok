package dynamicProxyController

type Config struct {
	IdentityPath  string               `df:"required"`
	ServiceName   string               `df:"required"`
	AmqpPublisher *AmqpPublisherConfig `df:"required"`
}
