package main

import (
	"bytes"
	"encoding/json"
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
		m.Method = "send-template"
		message := new(mandrill.Message)
		// message.Html = "<p>Example HTML content</p>"
		// message.Text = "Example text content"
		// message.Subject = "Example Subject"
		// message.FromEmail = "dev@wurkhappy.com"
		// message.FromName = "Test WH"
		message.To = []mandrill.To{{Email: "matt@wurkhappy.com", Name: "Matt"}}
		m.Args["message"] = message
		m.Args["template_name"] = "example"
		m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}
		resp, err := m.Send()
		var requestData map[string]interface{}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		data := buf.Bytes()
		json.Unmarshal(data, &requestData)
		log.Print(requestData)
		log.Print(err)
		log.Print(d.Body)

		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
