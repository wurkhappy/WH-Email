package handlers

import (
	// "bytes"
	"encoding/json"
	"github.com/wurkhappy/mandrill-go"
	// "log"
	"fmt"
	// "net/http"
	"strconv"
	"time"
)

func PaymentRequest(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Payment Request"
	vars := make([]*mandrill.GlobalVar, 0)
	err := paymentFreelancerSendToClient(body, template, vars)
	if err != nil {
		return err
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
	ID           string    `json:"id"`
	Amount       float64   `json:"amount"`
	Title        string    `json:"title"`
	DateExpected time.Time `json:"dateExpected"`
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
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: WebServerURI + path + "?" + signatureParams},
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
		return err
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
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: WebServerURI + path + "?" + signatureParams},
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
		return err
	}
	return nil

}
