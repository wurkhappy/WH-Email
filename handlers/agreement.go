package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/mandrill-go"
	"strconv"
)

func NewAgreement(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Agreement New User"
	vars := make([]*mandrill.GlobalVar, 0)
	err := agrmntFreelancerSendToClient(body, template, vars)
	if err != nil {
		return err
	}
	return nil
}

func AgreementChange(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Agreement Change"
	vars := make([]*mandrill.GlobalVar, 0)
	err := agrmntFreelancerSendToClient(body, template, vars)
	if err != nil {
		return err
	}
	return nil
}

func AgreementAccept(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Agreement Accept"
	vars := make([]*mandrill.GlobalVar, 0)
	err := agrmntClientSendToFreelancer(body, template, vars)
	if err != nil {
		return err
	}
	return nil
}

func AgreementReject(params map[string]string, body map[string]*json.RawMessage) error {
	template := "Agreement Dispute"
	vars := make([]*mandrill.GlobalVar, 0)
	err := agrmntClientSendToFreelancer(body, template, vars)
	if err != nil {
		return err
	}
	return nil
}

type Agreement struct {
	AgreementID  string     `json:"agreementID"`
	VersionID    string     `json:"versionID" bson:"_id"`
	Version      float64    `json:"version"`
	ClientID     string     `json:"clientID"`
	FreelancerID string     `json:"freelancerID"`
	Title        string     `json:"title"`
	Payments     []*Payment `json:"payments"`
}

func (a *Agreement) getTotalCost() float64 {
	var totalCost float64
	payments := a.Payments
	for _, payment := range payments {
		totalCost += payment.Amount
	}
	return totalCost
}

func agrmntClientSendToFreelancer(body map[string]*json.RawMessage, template string, vars []*mandrill.GlobalVar) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
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

func agrmntFreelancerSendToClient(body map[string]*json.RawMessage, template string, vars []*mandrill.GlobalVar) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
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
