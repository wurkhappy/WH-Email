package handlers

import (
	"bytes"
	// "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
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

type paymentAction struct {
	VersionID string `json:"versionID"`
	PaymentID string `json:"paymentID"`
	UserID    string `json:"userID"`
	Message   string `json:"message"`
}

func PaymentRequest(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var action *paymentAction
	json.Unmarshal(body, &action)
	agreement, payments, tasks := getAllInfo(action.VersionID)
	payment := payments.getPayment(action.PaymentID)

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := payment.LastAction.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, payments, tasks, action.Message, sender, recipient)
	data["Payment"] = payment

	dataSummary := map[string]interface{}{
		"agreement":  agreement,
		"tasks":      tasks,
		"payments":   payments,
		"payment":    payment,
		"freelancer": freelancer,
		"client":     client,
	}
	d, _ := json.Marshal(dataSummary)
	tplSummary, _ := sendServiceRequest("GET", config.PDFTemplatesService, "/template/invoice", d)
	pdfResp, _ := sendServiceRequest("POST", config.PDFService, "/string", tplSummary)

	var html bytes.Buffer
	paymentRequestTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.CC = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " requests payment"
	mail.Html = html.String()
	mail.Attachments = append(mail.Attachments, &models.Attachment{Type: "application/pdf", Name: "Invoice.pdf", Content: string(pdfResp)})

	_, err = mail.Send()
	return nil, err, 200
}

func PaymentAccepted(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var action *paymentAction
	json.Unmarshal(body, &action)
	agreement, payments, tasks := getAllInfo(action.VersionID)
	payment := payments.getPayment(action.PaymentID)

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := payment.LastAction.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, payments, tasks, action.Message, sender, recipient)

	var html bytes.Buffer
	paymentReceivedTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = sender.getEmailOrName() + " Just Paid You"
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}

func PaymentSent(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var action *paymentAction
	json.Unmarshal(body, &action)
	agreement, payments, tasks := getAllInfo(action.VersionID)
	payment := payments.getPayment(action.PaymentID)

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := payment.LastAction.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, payments, tasks, action.Message, sender, recipient)

	var html bytes.Buffer
	paymentSentTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = "You just paid " + recipient.getEmailOrName()
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200

}

func PaymentReject(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var action *paymentAction
	json.Unmarshal(body, &action)
	agreement, payments, tasks := getAllInfo(action.VersionID)
	payment := payments.getPayment(action.PaymentID)

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := payment.LastAction.WhoIsSenderRecipient(freelancer, client)

	data := createPaymentData(agreement, payment, payments, tasks, action.Message, sender, recipient)

	var html bytes.Buffer
	paymentDisputeTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = sender.getEmailOrName() + " Has Disputed Your Request"
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}

type Payment struct {
	ID           string       `json:"id"`
	VersionID    string       `json:"versionID"`
	Title        string       `json:"title"`
	DateExpected time.Time    `json:"dateExpected"`
	PaymentItems PaymentItems `json:"paymentItems"`
	LastAction   *Action      `json:"lastAction"`
	IsDeposit    bool         `json:"isDeposit"`
	AmountDue    float64      `json:"amountDue"`
	AmountPaid   float64      `json:"amountPaid"`
}

type Payments []*Payment

type PaymentItem struct {
	TaskID    string  `json:"taskID"`
	SubTaskID string  `json:"subtaskID"`
	Hours     float64 `json:"hours"`
	AmountDue float64 `json:"amountDue"`
	Rate      float64 `json:"rate"`
	Title     string  `json:"title"`
}

type PaymentItems []*PaymentItem

func (p Payments) getTotalCost() float64 {
	var totalCost float64
	for _, payment := range p {
		totalCost += payment.AmountDue
	}
	return totalCost
}

func (p Payments) getPayment(paymentID string) *Payment {
	for _, payment := range p {
		if payment.ID == paymentID {
			return payment
		}
	}
	return nil
}

func createPaymentData(agreement *Agreement, payment *Payment, payments Payments, tasks Tasks, message string, sender *User, recipient *User) map[string]interface{} {
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
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(tasks)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", payments.getTotalCost()),
		"PAYMENT_AMOUNT":         payment.AmountDue,
		"PAYMENT_REQUESTED_DATE": time.Now().Format("01/02/2006"),
		"WORK_ITEMS":             payment.PaymentItems,
	}
	return data
}
