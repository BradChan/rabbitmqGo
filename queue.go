/*
@Time : 2020/9/17 3:27 下午
@Author : chen
@File : queue
@Software: GoLand
*/
package rabbitmqGo

// QueueConfig is dedicated to consumer applications, not the producers.
type QueueConfig struct {
	// Name defines the name of the queue. Required.
	Name string
	// Binding defines the relationship between an exchange and a queue.
	// Optional.
	//
	// It is often used to refer to "routing key". Not required for the `fanout`
	// exchange types.
	Binding string
	// Exchange defines the name of the exchange that needs to be used for
	// message consumption. Required.
	//
	// This must match producer config value.
	Exchange string
	// Mode defines what type of queue shall be used. Required.
	//
	// This has an impact on the performance. Prefer `lazy` over `default`
	// unless you have a very reasonable case. Read reference below before
	// taking a decision.
	//
	// Ref: https://www.rabbitmq.com/lazy-queues.html
	Mode string
	// DLX is dedicated to the "dead-lettered" messages and represents the name
	// of the exchange that was declared previously. Optional.
	//
	// If the consumer does not/will never require a DLX feature, skip this
	// option. Late declaration of a DLX and programmatically using for an
	// existing queue is not possible. However, you can manually achieve this
	// which is not always wise as it is very error prone and tedious job.
	//
	// If the consumer requires a DLX feature, setup an exchange and queue
	// beforehand then use its name here. A DLX does not require its own
	// consumer upfront. It can be delivered when you know how to handle
	// "dead-lettered" messages.
	DLX string
}
