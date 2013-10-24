package handlers

import (
	// "bytes"
	// "encoding/json"
	"github.com/wurkhappy/mandrill-go"
	// "log"
	"fmt"
	// "net/http"
	"strconv"
	"time"
)

func PaymentRequest(params map[string]string, body map[string]interface{}) error {
	agreement := body["agreement"].(map[string]interface{})
	_ = body["payment"].(map[string]interface{})
	agreementID := agreement["id"].(string)
	clientID := agreement["clientID"].(string)
	freelancerID := agreement["freelancerID"].(string)
	userMessage := getUserMessage(body)
	payments := body["payments"].([]interface{})

	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(clientID, path, expiration)
	freelancerName := getEmailOrName(freelancer)
	totalCost := getTotalCost(agreement)

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement["title"].(string)},
		&mandrill.GlobalVar{Name: "FREELANCER_FULLNAME", Content: freelancerName},
		&mandrill.GlobalVar{Name: "CLIENT_MESSAGE", Content: userMessage},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: client["email"].(string), Name: createFullName(client)}}
	message.Subject = freelancerName + " Requests Payment"
	m.Args["message"] = message
	m.Args["template_name"] = "Payment Request"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func PaymentAccepted(params map[string]string, body map[string]interface{}) error {
	agreement := body["agreement"].(map[string]interface{})
	payment := body["payment"].(map[string]interface{})
	paymentAmount := payment["amount"].(float64)
	agreementID := agreement["id"].(string)
	clientID := agreement["clientID"].(string)
	freelancerID := agreement["freelancerID"].(string)
	_ = getUserMessage(body)
	payments := body["payments"].([]interface{})

	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(freelancerID, path, expiration)
	clientName := getEmailOrName(client)
	totalCost := getTotalCost(agreement)

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement["title"].(string)},
		&mandrill.GlobalVar{Name: "CLIENT_FULLNAME", Content: clientName},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: freelancer["email"].(string), Name: createFullName(freelancer)}}
	message.Subject = clientName + " Just Paid you $" + fmt.Sprintf("%g", paymentAmount)
	m.Args["message"] = message
	m.Args["template_name"] = "Payment Received"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func PaymentSent(params map[string]string, body map[string]interface{}) error {
	agreement := body["agreement"].(map[string]interface{})
	payment := body["payment"].(map[string]interface{})
	paymentAmount := payment["amount"].(float64)
	agreementID := agreement["id"].(string)
	clientID := agreement["clientID"].(string)
	freelancerID := agreement["freelancerID"].(string)
	_ = getUserMessage(body)
	payments := body["payments"].([]interface{})

	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(clientID, path, expiration)
	freelancerName := getEmailOrName(freelancer)
	totalCost := getTotalCost(agreement)

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement["title"].(string)},
		&mandrill.GlobalVar{Name: "FREELANCER_FIRSTNAME", Content: freelancerName},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: client["email"].(string), Name: createFullName(client)}}
	message.Subject = "You Just Paid " + freelancerName + " $" + fmt.Sprintf("%g", paymentAmount)
	m.Args["message"] = message
	m.Args["template_name"] = "Payment Sent"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func PaymentReject(params map[string]string, body map[string]interface{}) error {
	agreement := body["agreement"].(map[string]interface{})
	payment := body["payment"].(map[string]interface{})
	_ = payment["amount"].(float64)
	agreementID := agreement["id"].(string)
	clientID := agreement["clientID"].(string)
	freelancerID := agreement["freelancerID"].(string)
	userMessage := getUserMessage(body)
	payments := body["payments"].([]interface{})

	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(freelancerID, path, expiration)
	clientName := getEmailOrName(client)
	totalCost := getTotalCost(agreement)

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement["title"].(string)},
		&mandrill.GlobalVar{Name: "CLIENT_FULLNAME", Content: clientName},
		&mandrill.GlobalVar{Name: "CLIENT_MESSAGE", Content: userMessage},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: freelancer["email"].(string), Name: createFullName(freelancer)}}
	message.Subject = clientName + " has Disputed your Request"
	m.Args["message"] = message
	m.Args["template_name"] = "Payment Dispute"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}
