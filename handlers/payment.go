package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"github.com/wurkhappy/mandrill-go"
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

func PaymentRequest(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var payment *Payment
	json.Unmarshal(*body["payment"], &payment)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	data := createPaymentData(agreement, payment, message)

	var invoiceHTML bytes.Buffer
	invoiceTpl.Execute(&invoiceHTML, data)
	pdfResp, _ := sendServiceRequest("POST", config.PDFService, "/string", []byte(invoiceHTML.String()))
	attachment := base64.StdEncoding.EncodeToString(pdfResp)

	var html bytes.Buffer
	paymentRequestTpl.Execute(&html, data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	mail.FromEmail = "notifications@wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["FREELANCER_FULLNAME"].(string) + " requests payment"
	mail.Html = html.String()
	mail.Attachments = append(mail.Attachments, &models.Attachment{Type: "application/pdf", Name: "Invoice.pdf", Content: attachment})

	return mail.Send()
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

	data := createPaymentData(agreement, payment, message)

	var html bytes.Buffer
	paymentReceivedTpl.Execute(&html, data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["CLIENT_FULLNAME"].(string) + " Just Paid You"
	mail.Html = html.String()

	return mail.Send()
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

	data := createPaymentData(agreement, payment, message)

	var html bytes.Buffer
	paymentSentTpl.Execute(&html, data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = "You just paid " + data["FREELANCER_FULLNAME"].(string)
	mail.Html = html.String()

	return mail.Send()

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

	data := createPaymentData(agreement, payment, message)

	var html bytes.Buffer
	paymentDisputeTpl.Execute(&html, data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["CLIENT_FULLNAME"].(string) + " Has Disputed Your Request"
	mail.Html = html.String()

	return mail.Send()
}

type Payment struct {
	ID           string       `json:"id"`
	Amount       float64      `json:"amount"`
	Title        string       `json:"title"`
	ScopeItems   []*ScopeItem `json:"scopeItems"`
	DateExpected time.Time    `json:"dateExpected"`
}
type ScopeItem struct {
	Text string `json:"text"`
}

func paymentClientSendToFreelancer(body map[string]*json.RawMessage, template string, vars []*mandrill.GlobalVar) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var payment *Payment
	json.Unmarshal(*body["payment"], &payment)
	agreementID := agreement.VersionID
	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	path := "/agreement/v/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := 60 * 60 * 24
	signatureParams := createSignatureParams(freelancerID, path, expiration)
	clientName := client.getEmailOrName()
	totalCost := agreement.getTotalCost()

	var userMessage string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &userMessage)
	}

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: config.WebServer + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement.Title},
		&mandrill.GlobalVar{Name: "CLIENT_FULLNAME", Content: clientName},
		&mandrill.GlobalVar{Name: "MESSAGE", Content: userMessage},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(agreement.Payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
		&mandrill.GlobalVar{Name: "PAYMENT_AMOUNT", Content: fmt.Sprintf("%g", payment.Amount)},
	)
	for _, mergevar := range vars {
		message.GlobalMergeVars = append(message.GlobalMergeVars, mergevar)
	}
	message.To = []mandrill.To{{Email: freelancer.Email, Name: freelancer.createFullName()}}
	m.Args["message"] = message
	m.Args["template_name"] = template
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return fmt.Errorf("%s", err.Message)
	}
	return nil
}

func paymentFreelancerSendToClient(body map[string]*json.RawMessage, template string, vars []*mandrill.GlobalVar) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var payment *Payment
	json.Unmarshal(*body["payment"], &payment)
	agreementID := agreement.VersionID
	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	path := "/agreement/v/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := 60 * 60 * 24
	signatureParams := createSignatureParams(clientID, path, expiration)
	freelancerName := freelancer.getEmailOrName()
	totalCost := agreement.getTotalCost()

	var userMessage string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &userMessage)
	}

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: config.WebServer + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement.Title},
		&mandrill.GlobalVar{Name: "FREELANCER_FULLNAME", Content: freelancerName},
		&mandrill.GlobalVar{Name: "MESSAGE", Content: userMessage},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(agreement.Payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	for _, mergevar := range vars {
		message.GlobalMergeVars = append(message.GlobalMergeVars, mergevar)
	}
	message.To = []mandrill.To{{Email: client.Email, Name: client.createFullName()}}
	m.Args["message"] = message
	m.Args["template_name"] = template
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return fmt.Errorf("%s", err.Message)
	}
	return nil

}

func createPaymentData(agreement *Agreement, payment *Payment, message string) map[string]interface{} {
	agreementID := agreement.VersionID
	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	path := "/agreement/v/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := 60 * 60 * 24
	signatureParams := createSignatureParams(freelancerID, path, expiration)

	data := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"CLIENT_FULLNAME":        client.getEmailOrName(),
		"FREELANCER_FULLNAME":    freelancer.getEmailOrName(),
		"MESSAGE":                message,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.Payments)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", agreement.getTotalCost()),
		"PAYMENT_AMOUNT":         fmt.Sprintf("%g", payment.Amount),
		"PAYMENT_REQUESTED_DATE": time.Now().Format("01/02/2006"),
		"AGREEMENT_MILESTONE":    payment.Title,
		"WORK_ITEMS":             payment.ScopeItems,
		"CLIENT_EMAIL":           client.Email,
		"FREELANCER_EMAIL":       freelancer.Email,
	}
	return data
}
