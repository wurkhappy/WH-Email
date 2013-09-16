package main

import (
	"github.com/streadway/amqp"
	"github.com/wurkhappy/mandrill-go"
	"log"
)

func init() {
	mandrill.APIkey = "tKcqIfanhMnYrTtGrDixBA"
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		m := mandrill.NewCall()
		m.Category = "messages"
		m.Method = "send"
		message := map[string]interface{}{
			"template_name":    "example",
			"template_content": []mandrill.TemplateContent{{Name:"example", Content:"blah"}},
		}
		log.Print(message)

		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
