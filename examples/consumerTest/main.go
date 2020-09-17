/*
@Time : 2020/9/17 3:40 下午
@Author : chen
@File : main
@Software: GoLand
*/
package main

import (
	uuid "github.com/satori/go.uuid"
	"log"
	"rabbitmqGo"
	"rabbitmqGo/examples/consumer"
	"time"
)

func main() {
	log.Println("running keeper ...")

	// Consumers ---------------------------------------------------------------
	casillas0 := consumer.NewCasillas(rabbitmqGo.ConsumerConfig{
		Name:          "casillas",
		WorkerCount:   1,
		PrefetchCount: 3,
		ConsumerName:  "casillas0",
	})

	casillas1 := consumer.NewCasillas(rabbitmqGo.ConsumerConfig{
		Name:          "casillas",
		WorkerCount:   1,
		PrefetchCount: 3,
		ConsumerName:  "casillas1",
	})

	casillas2 := consumer.NewCasillas(rabbitmqGo.ConsumerConfig{
		Name:          "casillas",
		WorkerCount:   1,
		PrefetchCount: 3,
		ConsumerName:  "casillas2",
	})

	// RabbitMQ ----------------------------------------------------------------
	exchanges := []rabbitmqGo.ExchangeConfig{
		{
			Name: "penalty",
			Type: "direct",
		},
	}

	queues := []rabbitmqGo.QueueConfig{
		{
			Name:     "casillas",
			Binding:  "spain",
			Exchange: "penalty",
			Mode:     "lazy",
		},
	}

	logChan := make(chan rabbitmqGo.Log)
	go func() {
		log.Println("watching logs ...")
		for l := range logChan {
			log.Printf("%+v\n", l)
		}
	}()

	srv, err := rabbitmqGo.NewServer(rabbitmqGo.ConnectionConfig{
		Name:          "keeper",
		Schema:        "amqp",
		Username:      "user",
		Password:      "pwd",
		Host:          "www.hibottoy.com",
		Port:          "5672",
		VHost:         "ezaoyunmq",
		ReconAttempt:  300,
		ReconInterval: time.Second,
		CACert:        nil,
		ClientCert:    nil,
		ClientKey:     nil,
	}, logChan)
	if err != nil {
		log.Fatalln(err)
	}
	defer srv.Shutdown()

	if err := srv.Setup(exchanges, queues); err != nil {
		log.Fatalln(err)
	}

	if err := srv.RegisterConsumers([]rabbitmqGo.Consumer{casillas0, casillas1, casillas2}); err != nil {
		log.Fatalln(err)
	}

	select {}
}

func getUUID() string {
	uu := uuid.NewV1()
	return uu.String()
}
