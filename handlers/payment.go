package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/mandrill-go"
	"strconv"
	"time"
)

func PaymentRequest(params map[string]string, body map[string]*json.RawMessage) error {
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

	data := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"CLIENT_FULLNAME":        clientName,
		"FREELANCER_FULLNAME":    freelancer.getEmailOrName(),
		"MESSAGE":                userMessage,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.Payments)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", totalCost),
		"PAYMENT_AMOUNT":         fmt.Sprintf("%g", payment.Amount),
		"PAYMENT_REQUESTED_DATE": time.Now().Format("01/02/2006"),
		"AGREEMENT_MILESTONE":    payment.Title,
		"WORK_ITEMS":             payment.ScopeItems,
	}

	var invoiceHTML bytes.Buffer
	templates.ExecuteTemplate(&invoiceHTML, "invoice", data)
	pdfResp, _ := sendServiceRequest("POST", config.PDFService, "/string", []byte(invoiceHTML.String()))
	attachment := base64.StdEncoding.EncodeToString(pdfResp)

	var html bytes.Buffer
	templates.ExecuteTemplate(&html, "payment_request", data)

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send"
	message := new(mandrill.Message)
	message.To = []mandrill.To{{Email: client.Email, Name: client.createFullName()}}
	message.FromEmail = "notifications@wurkhappy.com"
	message.FromName = "Wurk Happy"
	message.Subject = freelancer.getEmailOrName() + " requests payment"
	message.Html = html.String()
	message.Attachments = append(message.Attachments, &mandrill.Attachment{Type: "application/pdf", Name: "Invoice.pdf", Content: attachment})
	m.Args["message"] = message

	_, err := m.Send()
	if err != nil {
		return fmt.Errorf("%s", err.Message)
	}
	return nil
}

func PaymentAccepted(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Payment Received"
	vars := make([]*mandrill.GlobalVar, 0)
	err := paymentClientSendToFreelancer(body, template, vars)
	if err != nil {
		return err
	}
	return nil
}

func PaymentSent(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Payment Sent"
	vars := make([]*mandrill.GlobalVar, 0)
	err := paymentClientSendToFreelancer(body, template, vars)
	if err != nil {
		return err
	}
	return nil
}

func PaymentReject(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Payment Dispute"
	vars := make([]*mandrill.GlobalVar, 0)
	err := paymentClientSendToFreelancer(body, template, vars)
	if err != nil {
		return err
	}
	return nil
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
