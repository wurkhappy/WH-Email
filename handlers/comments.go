package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/mandrill-go"
	"log"
	"time"
)

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
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.FromEmail = "reply-" + message_id.String() + "@notifications.wurkhappy.com"
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: config.WebServer + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement.Title},
		&mandrill.GlobalVar{Name: "SENDER_FULLNAME", Content: sender.getEmailOrName()},
		&mandrill.GlobalVar{Name: "MESSAGE", Content: comment.Text},
		&mandrill.GlobalVar{Name: "MESSAGE_ID", Content: message_id.String()},
	)
	message.To = []mandrill.To{{Email: recipient.Email, Name: recipient.createFullName()}}
	m.Args["message"] = message
	m.Args["template_name"] = "New Message"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return fmt.Errorf("%s", err.Message)
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
