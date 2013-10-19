package handlers

import (
	// "bytes"
	// "encoding/json"
	"github.com/wurkhappy/mandrill-go"
	// "log"
	"fmt"
	// "net/http"
	"strconv"
)

func init() {
	mandrill.APIkey = "AiZeQTNtBDY4omKvajApkg"
}

func ConfirmSignup(params map[string]string, body map[string]interface{}) error {
	userID := body["id"].(string)
	email := body["email"].(string)
	path := "/user/" + userID + "/verify"
	signature := signURL(userID, path)
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "signup_link", Content: "http://localhost:4000" + path + "?signature=" + signature + "&access_key=" + userID},
	)
	message.To = []mandrill.To{{Email: email, Name: createFullName(body)}}
	m.Args["message"] = message
	m.Args["template_name"] = "Confirm Email and Signup"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func NewAgreement(params map[string]string, body map[string]interface{}) error {
	agreementID := body["id"].(string)
	clientID := body["clientID"].(string)
	freelancerID := body["freelancerID"].(string)
	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	signature := signURL(clientID, path)
	freelancerName := createFullName(freelancer)
	if freelancerName == "" {
		freelancerName = freelancer["email"].(string)
	}
	var totalCost float64
	payments := body["payments"].([]interface{})
	for _, payment := range payments {
		model := payment.(map[string]interface{})
		totalCost += model["amount"].(float64)
	}

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?signature=" + signature + "&access_key=" + clientID},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: body["title"].(string)},
		&mandrill.GlobalVar{Name: "FREELANCER_NAME", Content: freelancerName},
		&mandrill.GlobalVar{Name: "PAYMENTS_TOTAL", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: client["email"].(string), Name: createFullName(client)}}
	message.Subject = freelancerName + " Sent you a New Agreeement"
	m.Args["message"] = message
	m.Args["subject"] = freelancerName + " Sent you a New Agreeement"
	m.Args["template_name"] = "Agreement New User"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func AgreementAccept(params map[string]string, body map[string]interface{}) error {
	agreement := body["agreement"].(map[string]interface{})
	agreementID := agreement["id"].(string)
	clientID := agreement["clientID"].(string)
	freelancerID := agreement["freelancerID"].(string)
	var userMessage string = " "
	if msg, ok := body["message"]; ok {
		userMessage = msg.(string)
	}

	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	signature := signURL(freelancerID, path)
	clientName := createFullName(client)
	if clientName == "" {
		clientName = client["email"].(string)
	}
	var totalCost float64
	payments := agreement["payments"].([]interface{})
	for _, payment := range payments {
		model := payment.(map[string]interface{})
		totalCost += model["amount"].(float64)
	}

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?signature=" + signature + "&access_key=" + freelancerID},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement["title"].(string)},
		&mandrill.GlobalVar{Name: "CLIENT_FULLNAME", Content: clientName},
		&mandrill.GlobalVar{Name: "CLIENT_MESSAGE", Content: userMessage},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: freelancer["email"].(string), Name: createFullName(freelancer)}}
	message.Subject = clientName + " accepted your agreement"
	m.Args["message"] = message
	m.Args["template_name"] = "Agreement Accept"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func AgreementReject(params map[string]string, body map[string]interface{}) error {
	agreement := body["agreement"].(map[string]interface{})
	agreementID := agreement["id"].(string)
	clientID := agreement["clientID"].(string)
	freelancerID := agreement["freelancerID"].(string)
	var userMessage string = " "
	if msg, ok := body["message"]; ok {
		userMessage = msg.(string)
	}

	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	signature := signURL(freelancerID, path)
	clientName := createFullName(client)
	if clientName == "" {
		clientName = client["email"].(string)
	}
	var totalCost float64
	payments := agreement["payments"].([]interface{})
	for _, payment := range payments {
		model := payment.(map[string]interface{})
		totalCost += model["amount"].(float64)
	}

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?signature=" + signature + "&access_key=" + freelancerID},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement["title"].(string)},
		&mandrill.GlobalVar{Name: "CLIENT_FULLNAME", Content: clientName},
		&mandrill.GlobalVar{Name: "CLIENT_MESSAGE", Content: userMessage},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: freelancer["email"].(string), Name: createFullName(freelancer)}}
	message.Subject = clientName + " has Disputed your Request"
	m.Args["message"] = message
	m.Args["template_name"] = "Agreement Dispute"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func AgreementChange(params map[string]string, body map[string]interface{}) error {
	agreement := body["agreement"].(map[string]interface{})
	agreementID := agreement["id"].(string)
	clientID := agreement["clientID"].(string)
	freelancerID := agreement["freelancerID"].(string)
	var userMessage string = " "
	if msg, ok := body["message"]; ok {
		userMessage = msg.(string)
	}

	path := "/agreement/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	signature := signURL(clientID, path)
	freelancerName := createFullName(freelancer)
	if freelancerName == "" {
		freelancerName = freelancer["email"].(string)
	}
	var totalCost float64
	payments := agreement["payments"].([]interface{})
	for _, payment := range payments {
		model := payment.(map[string]interface{})
		totalCost += model["amount"].(float64)
	}

	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "AGREEMENT_LINK", Content: "http://localhost:4000" + path + "?signature=" + signature + "&access_key=" + clientID},
		&mandrill.GlobalVar{Name: "AGREEMENT_NAME", Content: agreement["title"].(string)},
		&mandrill.GlobalVar{Name: "USER_FULLNAME", Content: freelancerName},
		&mandrill.GlobalVar{Name: "CLIENT_MESSAGE", Content: userMessage},
		&mandrill.GlobalVar{Name: "AGREEMENT_NUM_PAYMENTS", Content: strconv.Itoa(len(payments))},
		&mandrill.GlobalVar{Name: "AGREEMENT_COST", Content: fmt.Sprintf("%g", totalCost)},
	)
	message.To = []mandrill.To{{Email: client["email"].(string), Name: createFullName(client)}}
	message.Subject = freelancerName + " Requests Changes to Your Agreement"
	m.Args["message"] = message
	m.Args["template_name"] = "Agreement Change"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}


