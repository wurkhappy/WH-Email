package main

import (
	"encoding/json"
	"flag"
	"github.com/ant0ine/go-urlrouter"
	"github.com/streadway/amqp"
	rbtmq "github.com/wurkhappy/Rabbitmq-go-wrapper"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/handlers"
	"github.com/wurkhappy/WH-Email/models"
	"log"
)

var (
	production   = flag.Bool("production", false, "Production settings")
	staging      = flag.Bool("staging", false, "Production settings")
	exchangeType = flag.String("exchange-type", "topic", "Exchange type - direct|fanout|topic|x-custom")
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
		urlrouter.Route{
			PathExp: "/comment",
			Dest:    handlers.SendComment,
		},
		urlrouter.Route{
			PathExp: "/comment/reply",
			Dest:    handlers.ProcessReply,
		},
	},
}

func main() {
	flag.Parse()
	if *production {
		config.Prod()
	} else if *staging {
		config.Prod()
	} else {
		config.Test()
	}
	handlers.Setup(*production)
	models.Setup(*production)
	conn, err := amqp.Dial(config.EmailBroker)
	if err != nil {
		log.Printf("Dial: %s", err)
	}
	c, err := rbtmq.NewConsumer(conn, config.EmailExchange, *exchangeType, config.EmailQueue, *consumerTag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	deliveries := c.Consume(config.EmailQueue)

	err = router.Start()
	if err != nil {
		panic(err)
	}
	for d := range deliveries {
		go routeMapper(d)
	}
	log.Print("deliveries ended")
}

func routeMapper(d amqp.Delivery) {
	route, params, err := router.FindRoute(d.RoutingKey)
	if err != nil || route == nil {
		log.Printf("ERROR is: %v", err)
		return
	}

	var m map[string]*json.RawMessage
	json.Unmarshal(d.Body, &m)
	var body map[string]*json.RawMessage
	json.Unmarshal(*m["Body"], &body)
	handler := route.Dest.(func(map[string]string, map[string]*json.RawMessage) error)

	log.Println(d.RoutingKey, string(*m["Body"]))

	err = handler(params, body)
	if err != nil {
		log.Printf("ERROR is: %v", err)
		d.Nack(false, false)
		return
	}
	d.Ack(false)
}
