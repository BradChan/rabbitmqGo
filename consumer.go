/*
@Time : 2020/9/17 3:28 下午
@Author : chen
@File : consumer
@Software: GoLand
*/
package rabbitmqGo

import "github.com/streadway/amqp"

type ConsumerConfig struct {
	// Name is used to name the consumer workers. It provides visual cues to
	// which channel is used by which consumer/worker in Management Plugin. It
	// also appears in the logs. Required.
	Name string
	// WorkerCount helps running given amount of workers for the consumer.
	// Required.
	//
	// This has a high impact on the performance. The performance also has a
	// direct relationship with the `PrefetchCount` option. If you have a fairly
	// busy queue, avoid setting it to `1`. Also avoid setting it to very high
	// because the more workers, the more work the broker has to do to keep
	// track of them. Read reference below before taking a decision.
	//
	// Ref: https://www.rabbitmq.com/blog/2012/04/25/rabbitmq-performance-measurements-part-2/
	// 1 -> n receiving rate vs consumer count / prefetch count
	WorkerCount int
	// PrefetchCount helps defining how many messages should be delivered to a
	// consumer before acknowledgments are received. Optional.
	//
	// This has a high impact on the performance. The performance also has a
	// direct relationship with the `WorkerCount` option below. Unless you have
	// a fairly quiet queue, avoid setting it to `1`. Read reference below
	// before taking a decision. Optional.
	//
	// Ref: https://www.rabbitmq.com/blog/2012/04/25/rabbitmq-performance-measurements-part-2/
	// n -> 0 sending bytes rate vs number of producers, for various message sizes
	// 1 -> n receiving rate vs consumer count / prefetch count
	PrefetchCount int

	ConsumerName string
}

// All consumers must implement this interface.
type Consumer interface {
	Config() ConsumerConfig
	Consume(messages <-chan amqp.Delivery, workerID int)
}
