/*
@Time : 2020/9/17 3:28 下午
@Author : chen
@File : producer
@Software: GoLand
*/
package rabbitmqGo

type ProducerConfig struct {
	// Name is used to reserve channels per producers. Hence the reason, each
	// producer must have an unique name. Required.
	Name string
	// ExchangeName defines the name of the exchange that needs to be used for
	// message publishing. Required.
	//
	// This must match consumer config value.
	ExchangeName string
	// RoutingKey is defined on the message. Optional.
	//
	// When the message is published, it ends up in a queue whose binding key
	// matches to the routing key. Not required for the `fanout` exchange types.
	RoutingKey string
}

// All producers must implement this interface.
type Producer interface {
	Produce(messageID string, message []byte, data interface{}) error
}
