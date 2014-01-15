package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"math/rand"
	"strconv"
	"time"
)

var newAgreementTpl *template.Template
var agreementChangeTpl *template.Template
var agreementAcceptTpl *template.Template
var agreementDisputeTpl *template.Template
var agreementRequestTpl *template.Template
var agreementSentTpl *template.Template
var agreementVoidedTpl *template.Template
var agreementSummaryTpl *template.Template

func init() {
	formatDate := func(date time.Time) string {
		return date.Format("Jan 2, 2006")
	}
	var err error
	newAgreementTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_new_user.html",
	))
	agreementSentTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_sent.html",
	))
	agreementChangeTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_change.html",
	))
	agreementAcceptTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_accept.html",
	))
	agreementDisputeTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_dispute.html",
	))
	agreementRequestTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_request.html",
	))
	agreementSentTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_sent.html",
	))
	agreementVoidedTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_voided.html",
	))
	agreementSummaryTpl, err = template.New("agreement_summary.html").Funcs(template.FuncMap{"formatDate": formatDate, "unescape": unescaped}).ParseFiles(
		"templates/agreement_summary.html",
	)
	fmt.Println(err)
}

func NewAgreement(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	if !recipient.IsRegistered {
		newAgreementTpl.ExecuteTemplate(&html, "base", data)
	} else {
		agreementSentTpl.ExecuteTemplate(&html, "base", data)
	}

	var summaryHTML bytes.Buffer
	dataSummary := map[string]interface{}{
		"agreement":   agreement,
		"totalAmount": data["AGREEMENT_COST"],
	}
	err := agreementSummaryTpl.Execute(&summaryHTML, dataSummary)
	pdfResp, _ := sendServiceRequest("POST", config.PDFService, "/string", summaryHTML.Bytes())

	threadID := agreement.VersionID
	threadID += recipient.ID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = whName + agreement.VersionID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Has Just Sent You A New Agreement"
	mail.Html = html.String()
	mail.Attachments = append(mail.Attachments, &models.Attachment{Type: "application/pdf", Name: "Agreement.pdf", Content: string(pdfResp)})

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err
}

func AgreementChange(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	agreementChangeTpl.ExecuteTemplate(&html, "base", data)

	threadID := agreement.VersionID
	threadID += recipient.ID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = whName + agreement.VersionID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Requests Changes to Your Agreement"
	mail.Html = html.String()

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err
}

func AgreementAccept(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	agreementAcceptTpl.ExecuteTemplate(&html, "base", data)

	threadID := agreement.VersionID
	threadID += recipient.ID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = whName + agreement.VersionID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Accepted Your Agreement"
	mail.Html = html.String()

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err
}

func AgreementReject(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	agreementDisputeTpl.ExecuteTemplate(&html, "base", data)

	threadID := agreement.VersionID
	threadID += recipient.ID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = whName + agreement.VersionID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Has Disputed Your Request"
	mail.Html = html.String()

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err
}

type Agreement struct {
	AgreementID         string     `json:"agreementID"`
	VersionID           string     `json:"versionID" bson:"_id"`
	Version             float64    `json:"version"`
	ClientID            string     `json:"clientID"`
	FreelancerID        string     `json:"freelancerID"`
	Title               string     `json:"title"`
	Payments            []*Payment `json:"payments"`
	WorkItems           WorkItems  `json:"workItems"`
	DraftCreatorID      string     `json:"draftCreatorID"`
	CurrentStatus       *Status    `json:"currentStatus"`
	ProposedServices    string     `json:"proposedServices"`
	AcceptsCreditCard   bool       `json:"acceptsCreditCard"`
	AcceptsBankTransfer bool       `json:"acceptsBankTransfer"`
}

type Status struct {
	ID                 string    `json:"id" bson:"_id"`
	AgreementID        string    `json:"agreementID"`
	AgreementVersionID string    `json:"agreementVersionID"`
	AgreementVersion   int       `json:"agreementVersion"`
	ParentID           string    `json:"parentID"`
	PaymentID          string    `json:"paymentID"`
	Action             string    `json:"action"`
	Date               time.Time `json:"date"`
	UserID             string    `json:"userID"`
}

func (s *Status) WhoIsSenderRecipient(user1 *User, user2 *User) (sender *User, recipient *User) {
	if s.UserID == user1.ID {
		return user1, user2
	}
	return user2, user1
}

func (a *Agreement) getTotalCost() float64 {
	var totalCost float64
	workItems := a.WorkItems
	for _, workItem := range workItems {
		totalCost += workItem.Amount
	}
	return totalCost
}

func createAgreementData(agreement *Agreement, message string, sender *User, recipient *User) (data map[string]interface{}) {
	agreementID := agreement.VersionID
	path := "/agreement/v/" + agreementID
	expiration := 60 * 60 * 24 * 7 * 4
	signatureParams := createSignatureParams(recipient.ID, path, expiration, recipient.IsVerified)

	m := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"SENDER_FULLNAME":        sender.getEmailOrName(),
		"RECIPIENT_FULLNAME":     recipient.getEmailOrName(),
		"MESSAGE":                message,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.WorkItems)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", agreement.getTotalCost()),
	}
	return m
}
