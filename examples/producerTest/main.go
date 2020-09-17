/*
@Time : 2020/9/17 3:40 下午
@Author : chen
@File : main
@Software: GoLand
*/
package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"log"
	"rabbitmqGo"
	"rabbitmqGo/examples/producer"
	"time"
)

func main() {
	log.Println("running striker ...")

	// RabbitMQ ----------------------------------------------------------------
	exchanges := []rabbitmqGo.ExchangeConfig{
		{
			Name: "penalty",
			Type: "direct",
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
		Name:          "striker",
		Schema:        "amqp",
		Username:      "user",
		Password:      "leban",
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

	if err := srv.Setup(exchanges, nil); err != nil {
		log.Fatalln(err)
	}

	// Producers ---------------------------------------------------------------
	morientes := producer.NewMorientes(srv, rabbitmqGo.ProducerConfig{
		ExchangeName: "penalty",
		RoutingKey:   "spain",
	})
	//morientes1 := producer.NewMorientes(srv, rabbitmqGo.ProducerConfig{
	//	ExchangeName: "penalty",
	//	RoutingKey:   "spain",
	//})

	uu := uuid.NewV1()
	//err = morientes.Produce(uu.String(),[]byte(fmt.Sprint(time.Now())),nil)
	err = morientes.Produce(uu.String(), []byte(uuid.NewV1().String()), nil)
	err = morientes.Produce(uu.String(), []byte(uuid.NewV1().String()), nil)
	err = morientes.Produce(uu.String(), []byte(uuid.NewV1().String()), nil)
	//err = morientes1.Produce(uu.String(),[]byte(uu.String()),nil)
	if err != nil {
		fmt.Println(err)
	}
	//morientes.Produce(uu.String(),[]byte(fmt.Sprint(time.Now())),nil)
	//morientes.Produce(uu.String(),[]byte(fmt.Sprint(time.Now())),nil)
	// ....
}
