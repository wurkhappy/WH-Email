package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"math/rand"
	"strings"
	"time"
)

var newMessageTpl *template.Template

func init() {
	newMessageTpl = template.Must(template.ParseFiles(
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
	agreement := getAgreement(comment.AgreementID)
	var recipientID string = agreement.ClientID
	if agreement.ClientID == comment.UserID {
		recipientID = agreement.FreelancerID
	}
	recipient := getUserInfo(recipientID)
	comment.RecipientID = recipientID

	path := "/agreement/v/" + agreement.VersionID
	expiration := int(time.Now().Add(time.Hour * 24 * 7 * 4).Unix())
	signatureParams := createSignatureParams(recipientID, path, expiration, recipient.IsVerified)

	data := map[string]interface{}{
		"AGREEMENT_LINK":  config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":  agreement.Title,
		"SENDER_FULLNAME": sender.getEmailOrName(),
		"MESSAGE":         template.HTML(comment.Text),
		"RANDOM":          rand.Intn(8),
	}
	var html bytes.Buffer
	newMessageTpl.ExecuteTemplate(&html, "body", data)

	//join all tag IDs into a string to create a unique ID for threading
	var tagsJoined string
	var tagsSubject string
	tagsLength := len(comment.Tags)
	for i, tag := range comment.Tags {
		tagsJoined += tag.ID
		if tagsLength-1 == i && tagsLength > 1 {
			tagsSubject += " and " + tag.Name
		} else if i != 0 {
			tagsSubject += ", " + tag.Name
		} else {
			tagsSubject = strings.ToUpper(tag.Name[0:1]) + tag.Name[1:len(tag.Name)]
		}

	}
	if tagsSubject == "" {
		tagsSubject = agreement.Title
	}
	//add part of the user's ID so that it's unique for the user
	tagsJoined += recipient.ID[0:4]
	tagsJoined += comment.AgreementID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(tagsJoined, c)

	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.createFullName()}}
	mail.FromEmail = whName + tagsJoined[0:2] + "@notifications.wurkhappy.com"
	mail.Subject = tagsSubject
	mail.Html = html.String()

	msgID, err := mail.Send()
	if err != nil {
		return err
	}

	if threadMsgID == "" {
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return nil
}
