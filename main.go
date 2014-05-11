package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/handlers"
	"log"
)

var count int

var conn *amqp.Connection

var production = flag.Bool("production", false, "Production settings")

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func main() {
	var err error

	flag.Parse()
	if *production {
		config.Prod()
	} else {
		config.Test()
	}
	handlers.Setup(*production)

	err = router.Start()
	if err != nil {
		panic(err)
	}

	conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		config.MainExchange, // name
		"topic",             // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // noWait
		nil,                 // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	q, err := ch.QueueDeclare(
		config.EmailQueue, // name
		false,             // durable
		false,             // delete when usused
		false,             // exclusive
		false,             // noWait
		amqp.Table{
			"x-dead-letter-exchange": config.DeadLetterExchange,
		}, // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for _, route := range router.Routes {
		err = ch.QueueBind(
			q.Name,              // queue name
			route.PathExp,       // routing key
			config.MainExchange, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	for d := range msgs {
		routeDelivery(d)
	}
}

func routeDelivery(d amqp.Delivery) {
	route, _, err := router.FindRoute(d.RoutingKey)
	if err != nil || route == nil {
		log.Printf("ERROR is: %v", err)
		d.Nack(false, false)
		return
	}
	log.Println(d.RoutingKey)

	params := make(map[string]interface{})

	handler := route.Dest.(func(map[string]interface{}, []byte) ([]byte, error, int))
	_, err, _ = handler(params, d.Body)
	if err != nil {
		log.Printf("ERROR is: %v", err)
		d.Nack(false, false)
		return
	}
	d.Ack(false)
}
