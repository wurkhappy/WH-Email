package models

import (
	"github.com/wurkhappy/mandrill-go"
	"fmt"
)

type Mail struct {
	Html        string        `json:"html,omitempty"`
	Subject     string        `json:"subject"`
	FromEmail   string        `json:"from_email"`
	FromName    string        `json:"from_name"`
	To          []To          `json:"to"`
	BCCAddress  string        `json:"bcc_address,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type To struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (mail *Mail) Send() error {
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send"
	message := new(mandrill.Message)
	for _, to := range mail.To {
		message.To = append(message.To, mandrill.To{Email: to.Email, Name: to.Name})
	}
	message.FromEmail = mail.FromEmail
	message.FromName = mail.FromName
	message.Subject = mail.Subject
	message.Html = mail.Html
	for _, attachment := range mail.Attachments {
		message.Attachments = append(message.Attachments, &mandrill.Attachment{Type: attachment.Type, Name: attachment.Name, Content: attachment.Content})
	}
	m.Args["message"] = message

	_, err := m.Send()
	if err != nil {
		return fmt.Errorf("%s", err.Message)
	}
	return nil
}
