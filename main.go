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
			PathExp: "/user/password/forgot",
			Dest:    handlers.ForgotPassword,
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
		urlrouter.Route{
			PathExp: "/agreement/updated",
			Dest:    handlers.AgreementChange,
		},
		urlrouter.Route{
			PathExp: "/payment/submitted",
			Dest:    handlers.PaymentRequest,
		},
		urlrouter.Route{
			PathExp: "/payment/accepted",
			Dest:    handlers.PaymentAccepted,
		},
		urlrouter.Route{
			PathExp: "/payment/rejected",
			Dest:    handlers.PaymentReject,
		},
		urlrouter.Route{
			PathExp: "/payment/sent",
			Dest:    handlers.PaymentSent,
		},
		urlrouter.Route{
			PathExp: "/test",
			Dest:    handlers.Test,
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
	for d := range deliveries {
		go routeMapper(d)
	}
}

func routeMapper(d amqp.Delivery) {
	route, params, err := router.FindRoute(d.RoutingKey)
	if err != nil || route == nil {
		log.Printf("first error is: %v", err)
		return
	}

	var m map[string]*json.RawMessage
	json.Unmarshal(d.Body, &m)
	var body map[string]*json.RawMessage
	json.Unmarshal(*m["Body"], &body)
	handler := route.Dest.(func(map[string]string, map[string]*json.RawMessage) error)
	err = handler(params, body)
	if err != nil {
		log.Printf("second error is: %v", err)
		d.Nack(false, false)
	}
	d.Ack(false)
}
