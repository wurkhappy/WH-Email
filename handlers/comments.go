package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"log"
	"time"
)

var newMessageTpl *template.Template

func init() {
	newMessageTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/new_message.html",
	))
}

type Comment struct {
	ID                 string    `json:"id,omitempty"`
	UserID             string    `json:"userID"`
	RecipientID        string    `json:"recipientID"` //not part of the original model but we need this info for the reply email
	AgreementID        string    `json:"agreementID"`
	AgreementVersionID string    `json:"agreementVersionID"`
	DateCreated        time.Time `json:"dateCreated"`
	Text               string    `json:"text"`
	Tags               []*Tag    `json:"tags"`
}

type Tag struct {
	ID          string `json:"id"`
	AgreementID string `json:"agreementID"`
	Name        string `json:"name"`
}

func SendComment(params map[string]string, body map[string]*json.RawMessage) error {
	var comment *Comment
	json.Unmarshal(*body["comment"], &comment)
	sender := getUserInfo(comment.UserID)
	agreement := getAgreementOwners(comment.AgreementID)
	var recipientID string = agreement.ClientID
	if agreement.ClientID == comment.UserID {
		recipientID = agreement.FreelancerID
	}
	recipient := getUserInfo(recipientID)

	path := "/agreement/v/" + comment.AgreementVersionID
	expiration := int(time.Now().Add(time.Hour * 24 * 7 * 4).Unix())
	signatureParams := createSignatureParams(recipientID, path, expiration)

	data := map[string]interface{}{
		"AGREEMENT_LINK":  config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":  agreement.Title,
		"SENDER_FULLNAME": sender.getEmailOrName(),
		"MESSAGE":         template.HTML(comment.Text),
	}
	var html bytes.Buffer
	newMessageTpl.ExecuteTemplate(&html, "base", data)

	var tagsJoined string
	for _, tag := range comment.Tags {
		tagsJoined += tag.ID
	}
	fmt.Println(tagsJoined)
	c := redisPool.Get()
	replyTo, _ := redis.String(c.Do("GET", tagsJoined))

	mail := new(models.Mail)
	if replyTo != "" {
		mail.InReplyTo = replyTo
	}
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.createFullName()}}
	mail.FromEmail = "reply" + tagsJoined[0:2] + "@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = sender.getEmailOrName() + " Has Just Sent You A New Message"
	mail.Html = html.String()

	msgID, err := mail.Send()
	if err != nil {
		return err
	}
	fmt.Println(msgID)
	comment.RecipientID = recipientID
	jsonComment, _ := json.Marshal(comment)
	if _, err := c.Do("SET", msgID, jsonComment); err != nil {
		log.Panic(err)
	}
	if _, err := c.Do("SET", tagsJoined, msgID); err != nil {
		log.Panic(err)
	}
	return nil
}
