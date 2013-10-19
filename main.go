package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ant0ine/go-urlrouter"
	"github.com/streadway/amqp"
	rbtmq "github.com/wurkhappy/Rabbitmq-go-wrapper"
	"github.com/wurkhappy/WH-Email/handlers"
	"log"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchange     = flag.String("exchange", "email", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue        = flag.String("queue", "email", "Ephemeral AMQP queue name")
	consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
)

//order matters so most general should go towards the bottom
var router urlrouter.Router = urlrouter.Router{
	Routes: []urlrouter.Route{
		urlrouter.Route{
			PathExp: "/user/verify",
			Dest:    handlers.ConfirmSignup,
		},
		urlrouter.Route{
			PathExp: "/agreement/submitted",
			Dest:    handlers.NewAgreement,
		},
		urlrouter.Route{
			PathExp: "/agreement/accepted",
			Dest:    handlers.AgreementAccept,
		},
		urlrouter.Route{
			PathExp: "/agreement/rejected",
			Dest:    handlers.AgreementReject,
		},
	},
}

func init() {
	flag.Parse()
}

func main() {
	log.Printf("dialing %q", *uri)
	conn, err := amqp.Dial(*uri)
	if err != nil {
		fmt.Errorf("Dial: %s", err)
	}
	c, err := rbtmq.NewConsumer(conn, *exchange, *exchangeType, *queue, *consumerTag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	deliveries := c.Consume(*queue)

	err = router.Start()
	if err != nil {
		panic(err)
	}
	log.Print("route")
	routeMapper(deliveries)
	select {}
}

func routeMapper(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		//this should use a goroutine but channel closes under heavy load
		func(amqp.Delivery) {
			route, params, err := router.FindRoute(d.RoutingKey)
			if err != nil || route == nil {
				log.Printf("route is: ", route)
				log.Printf("first error is: %v", err)
				return
			}

			var m map[string]interface{}
			json.Unmarshal(d.Body, &m)
			body := m["Body"].(map[string]interface{})
			handler := route.Dest.(func(map[string]string, map[string]interface{}) error)
			err = handler(params, body)
			if err != nil {
				log.Printf("second error is: %v", err)
				d.Nack(false, true)
			}
			d.Ack(false)
		}(d)

	}
	log.Printf("handle: deliveries channel closed")
}
