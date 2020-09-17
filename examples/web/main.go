/*
@Time : 2020/9/17 4:52 下午
@Author : chen
@File : main
@Software: GoLand
*/
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
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

	if err := srv.Setup(exchanges, nil); err != nil {
		log.Fatalln(err)
	}

	r := gin.Default()
	r.GET("/info", func(c *gin.Context) {
		msg := c.Query("msg")

		produe := producer.NewMorientes(srv, rabbitmqGo.ProducerConfig{
			ExchangeName: "penalty",
			RoutingKey:   "spain",
		})
		err := produe.Produce(uuid.NewV1().String(), []byte(msg), nil)
		if err != nil {
			log.Fatalln(err)
			c.JSON(200, gin.H{
				"message": fmt.Sprint(err),
			})
		} else {
			c.JSON(200, gin.H{
				"message": msg,
			})
		}
	})

	r.Run("127.0.0.1:8071")

}
