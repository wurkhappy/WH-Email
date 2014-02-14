package handlers

import (
	"bytes"
	// "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"math/rand"
	"strconv"
	"time"
)

var paymentRequestTpl *template.Template
var invoiceTpl *template.Template
var paymentReceivedTpl *template.Template
var paymentSentTpl *template.Template
var paymentDisputeTpl *template.Template

func init() {
	paymentRequestTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/payment_request.html",
		"templates/invoice.html",
		"templates/invoice_fixed.html",
		"templates/invoice_hourly.html",
	))
	invoiceTpl = template.Must(template.ParseFiles(
		"templates/invoice.html",
		"templates/invoice_hourly.html",
		"templates/invoice_fixed.html",
	))
	paymentReceivedTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/payment_received.html",
	))
	paymentSentTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/payment_sent.html",
	))
	paymentDisputeTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/payment_dispute.html",
	))
}

//

func PaymentRequest(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var payment *Payment
	json.Unmarshal(*body["payment"], &payment)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}
	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)

	sender, recipient := payment.CurrentStatus.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, message, sender, recipient)
	data["Payment"] = payment

	var invoiceHTML bytes.Buffer
	if len(payment.PaymentItems) > 0 && payment.PaymentItems[0].Hours != 0 {
		data["hourly"] = true
	}
	invoiceTpl.ExecuteTemplate(&invoiceHTML, "invoice", data)
	pdfResp, _ := sendServiceRequest("POST", config.PDFService, "/string", []byte(invoiceHTML.String()))

	var html bytes.Buffer
	paymentRequestTpl.ExecuteTemplate(&html, "base", data)

	//this creates an ID for the message thread
	threadID := payment.ID
	threadID += recipient.ID[0:4]

	//use the thread ID to save info about the message which we will later use in order to continue the thread
	//we also use this in order to keep tags consistent across reply messages
	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}

	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = whName + payment.ID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " requests payment"
	mail.Html = html.String()
	mail.Attachments = append(mail.Attachments, &models.Attachment{Type: "application/pdf", Name: "Invoice.pdf", Content: string(pdfResp)})

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err
}

func PaymentAccepted(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var payment *Payment
	json.Unmarshal(*body["payment"], &payment)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)

	sender, recipient := payment.CurrentStatus.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, message, sender, recipient)

	var html bytes.Buffer
	paymentReceivedTpl.ExecuteTemplate(&html, "base", data)

	threadID := payment.ID
	threadID += recipient.ID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}

	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = whName + payment.ID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = sender.getEmailOrName() + " Just Paid You"
	mail.Html = html.String()

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err
}

func PaymentSent(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var payment *Payment
	json.Unmarshal(*body["payment"], &payment)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)

	sender, recipient := payment.CurrentStatus.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, message, recipient, sender)

	var html bytes.Buffer
	paymentSentTpl.ExecuteTemplate(&html, "base", data)

	threadID := payment.ID
	threadID += sender.ID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}

	mail.To = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = whName + payment.ID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = "You just paid " + recipient.getEmailOrName()
	mail.Html = html.String()

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err

}

func PaymentReject(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var payment *Payment
	json.Unmarshal(*body["payment"], &payment)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)

	sender, recipient := payment.CurrentStatus.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, message, sender, recipient)

	var html bytes.Buffer
	paymentDisputeTpl.ExecuteTemplate(&html, "base", data)

	threadID := payment.ID
	threadID += recipient.ID[0:4]

	c := redisPool.Get()
	threadMsgID := getThreadMessageID(threadID, c)
	mail := new(models.Mail)
	if threadMsgID != "" {
		mail.InReplyTo = threadMsgID
	}

	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = whName + payment.ID[0:rand.Intn(8)] + "@notifications.wurkhappy.com"
	mail.Subject = sender.getEmailOrName() + " Has Disputed Your Request"
	mail.Html = html.String()

	msgID, err := mail.Send()
	if threadMsgID == "" {
		comment := new(Comment)
		comment.AgreementID = agreement.AgreementID
		saveMessageInfo(threadMsgID, msgID, comment, sender, recipient, c)
	}
	return err
}

type Payment struct {
	ID            string       `json:"id"`
	Title         string       `json:"title"`
	DateExpected  time.Time    `json:"dateExpected"`
	PaymentItems  PaymentItems `json:"paymentItems"`
	CurrentStatus *Status      `json:"currentStatus"`
	IsDeposit     bool         `json:"isDeposit"`
	AmountDue     float64      `json:"amountDue"`
	AmountPaid    float64      `json:"amountPaid"`
}
type Payments []*Payment

type PaymentItem struct {
	WorkItemID string  `json:"workItemID"`
	TaskID     string  `json:"taskID"`
	Hours      float64 `json:"hours"`
	AmountDue  float64 `json:"amountDue"`
	Rate       float64 `json:"rate"`
	Title      string  `json:"title"`
}

type PaymentItems []*PaymentItem

type Task struct {
	ID           string    `json:"id"`
	Completed    bool      `json:"completed"`
	IsPaid       bool      `json:"isPaid"`
	Hours        float64   `json:"hours"`
	Tasks        Tasks     `json:"scopeItems"`
	Title        string    `json:"title"`
	DateExpected time.Time `json:"dateExpected"`
	Description  string    `json:"description"`
}
type Tasks []*Task

type ScopeItem struct {
	Text string `json:"text"`
}

func createPaymentData(agreement *Agreement, payment *Payment, message string, sender *User, recipient *User) map[string]interface{} {
	agreementID := agreement.VersionID
	path := "/agreement/v/" + agreementID
	expiration := 60 * 60 * 24 * 7 * 4
	signatureParams := createSignatureParams(recipient.ID, path, expiration, recipient.IsVerified)

	data := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"SENDER_FULLNAME":        sender.getEmailOrName(),
		"RECIPIENT_FULLNAME":     recipient.getEmailOrName(),
		"MESSAGE":                message,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.Tasks)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", agreement.getTotalCost()),
		"PAYMENT_AMOUNT":         payment.AmountDue,
		"PAYMENT_REQUESTED_DATE": time.Now().Format("01/02/2006"),
		"WORK_ITEMS":             payment.PaymentItems,
	}
	return data
}
