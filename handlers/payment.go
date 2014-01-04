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
	))
	invoiceTpl = template.Must(template.ParseFiles(
		"templates/invoice.html",
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

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data := createPaymentData(agreement, payment, message, sendToFreelancer)

	var invoiceHTML bytes.Buffer
	invoiceTpl.ExecuteTemplate(&invoiceHTML, "invoice", data)
	pdfResp, _ := sendServiceRequest("POST", config.PDFService, "/string", []byte(invoiceHTML.String()))

	var html bytes.Buffer
	paymentRequestTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	mail.FromEmail = "notifications@wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["FREELANCER_FULLNAME"].(string) + " requests payment"
	mail.Html = html.String()
	mail.Attachments = append(mail.Attachments, &models.Attachment{Type: "application/pdf", Name: "Invoice.pdf", Content: string(pdfResp)})

	_, err := mail.Send()
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

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data := createPaymentData(agreement, payment, message, sendToFreelancer)

	var html bytes.Buffer
	paymentReceivedTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["CLIENT_FULLNAME"].(string) + " Just Paid You"
	mail.Html = html.String()

	_, err := mail.Send()
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

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data := createPaymentData(agreement, payment, message, sendToFreelancer)

	var html bytes.Buffer
	paymentSentTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = "You just paid " + data["FREELANCER_FULLNAME"].(string)
	mail.Html = html.String()

	_, err := mail.Send()
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

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data := createPaymentData(agreement, payment, message, sendToFreelancer)

	var html bytes.Buffer
	paymentDisputeTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["CLIENT_FULLNAME"].(string) + " Has Disputed Your Request"
	mail.Html = html.String()

	_, err := mail.Send()
	return err
}

type Payment struct {
	ID              string       `json:"id"`
	PaymentItems    PaymentItems `json:"paymentItems"`
	CurrentStatus   *Status      `json:"currentStatus"`
	IncludesDeposit bool         `json:"includesDeposit"`
	DateCreated     time.Time    `json:"dateCreated"`
}

type PaymentItem struct {
	WorkItemID    string  `json:"workItemID"`
	Amount        float64 `json:"amount"`
	WorkItemTitle string  `json:"workItemTitle"`
}

type PaymentItems []*PaymentItem

type WorkItem struct {
	ID           string       `json:"id"`
	Amount       float64      `json:"amountDue"`
	Title        string       `json:"title"`
	ScopeItems   []*ScopeItem `json:"scopeItems"`
	DateExpected time.Time    `json:"dateExpected"`
}
type WorkItems []*WorkItem

func (w WorkItems) GetWorkItem(id string) *WorkItem {
	for _, workItem := range w {
		if workItem.ID == id {
			return workItem
		}
	}
	return nil
}

type ScopeItem struct {
	Text string `json:"text"`
}

func createPaymentData(agreement *Agreement, payment *Payment, message string, toFreelancer bool) map[string]interface{} {
	agreementID := agreement.VersionID
	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	path := "/agreement/v/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := 60 * 60 * 24 * 7 * 4
	var signatureParams string
	if toFreelancer {
		signatureParams = createSignatureParams(freelancerID, path, expiration)
	} else {
		signatureParams = createSignatureParams(clientID, path, expiration)
	}

	var paymentAmount float64
	for _, paymentItem := range payment.PaymentItems {
		paymentAmount += paymentItem.Amount
		workItem := agreement.WorkItems.GetWorkItem(paymentItem.WorkItemID)
		paymentItem.WorkItemTitle = workItem.Title
	}

	data := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"CLIENT_FULLNAME":        client.getEmailOrName(),
		"FREELANCER_FULLNAME":    freelancer.getEmailOrName(),
		"MESSAGE":                message,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.WorkItems)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", agreement.getTotalCost()),
		"PAYMENT_AMOUNT":         fmt.Sprintf("%g", paymentAmount),
		"PAYMENT_REQUESTED_DATE": time.Now().Format("01/02/2006"),
		"WORK_ITEMS":             payment.PaymentItems,
		"CLIENT_EMAIL":           client.Email,
		"FREELANCER_EMAIL":       freelancer.Email,
	}
	return data
}
