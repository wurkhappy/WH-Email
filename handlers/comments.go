package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nu7hatch/gouuid"
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
	message_id, _ := uuid.NewV4()
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
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(recipientID, path, expiration)

	data := map[string]interface{}{
		"AGREEMENT_LINK":  config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":  agreement.Title,
		"SENDER_FULLNAME": sender.getEmailOrName(),
		"MESSAGE":         template.HTML(comment.Text),
		"MESSAGE_ID":      message_id.String(),
	}
	var html bytes.Buffer
	newMessageTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.createFullName()}}
	mail.FromEmail = "reply-" + message_id.String() + "@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = sender.getEmailOrName() + " Has Just Sent You A New Message"
	mail.Html = html.String()

	err := mail.Send()
	if err != nil {
		return err
	}
	c := redisPool.Get()
	fmt.Println(message_id.String())
	comment.RecipientID = recipientID
	jsonComment, _ := json.Marshal(comment)
	if _, err := c.Do("SET", message_id.String(), jsonComment); err != nil {
		log.Panic(err)
	}
	return nil
}
