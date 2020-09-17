/*
@Time : 2020/9/17 3:37 下午
@Author : chen
@File : casillas
@Software: GoLand
*/
package consumer

import (
	"github.com/streadway/amqp"
	"log"
	"rabbitmqGo"
)

type Casillas struct {
	config rabbitmqGo.ConsumerConfig
}

func NewCasillas(config rabbitmqGo.ConsumerConfig) Casillas {
	return Casillas{config: config}
}

func (c Casillas) Config() rabbitmqGo.ConsumerConfig {
	return c.config
}

func (c Casillas) Consume(messages <-chan amqp.Delivery, id int) {
	for message := range messages {
		// Do the work ...
		//log.Printf("[%d] %s consumed: %s\n", id, c.config.Name, string(message.Body))
		log.Printf("[%d] %s consumed: %s\n", id, c.config.ConsumerName, string(message.Body))

		if err := message.Ack(false); err != nil {
			log.Printf("consume: ack message: %v\n", err)
		}
	}
}
