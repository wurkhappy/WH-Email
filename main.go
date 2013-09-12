package main

import (
	"log"
	"net/smtp"
)

func main() {
	sender := "contact@wurkhappy.com"
	recipient := "example@test.com"
	sub := "Subject: Test\r\n\r\n"
	content := "Body"
	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		sender,
		"password",
		"smtp.gmail.com",
	)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		sender,
		[]string{recipient},
		[]byte(sub+content),
	)
	if err != nil {
		log.Fatal(err)
	}
}
