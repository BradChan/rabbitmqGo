/*
@Time : 2020/9/17 3:36 下午
@Author : chen
@File : morientes
@Software: GoLand
*/
package producer

import (
	"github.com/streadway/amqp"
	"rabbitmqGo"
)

type Morientes struct {
	rabbitmq *rabbitmqGo.Server
	config   rabbitmqGo.ProducerConfig
}

func NewMorientes(rabbitmq *rabbitmqGo.Server, config rabbitmqGo.ProducerConfig) Morientes {
	return Morientes{rabbitmq: rabbitmq, config: config}
}

func (m Morientes) Produce(messageID string, message []byte, data interface{}) error {
	// if err := m.rabbitmq.PublishOnReservedChannel(amqp.Publishing{
	if err := m.rabbitmq.PublishOnNewChannel(amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		ContentType:     "text/plain",
		ContentEncoding: "utf-8",
		MessageId:       messageID,
		Body:            message,
	}, m.config); err != nil {
		return err
	}

	return nil
}
