package dynamicProxy

const V = 1

type Config struct {
	V              int                   `df:"required"`
	AmqpSubscriber *AmqpSubscriberConfig `df:"required"`
}
