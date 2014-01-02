package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
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

func init() {
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
}

func NewAgreement(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data, _, client := createAgreementData(agreement, message, sendToFreelancer)

	mail := new(models.Mail)
	status := agreement.CurrentStatus
	if status.UserID == agreement.FreelancerID {
		data["SENDER_FULLNAME"] = data["FREELANCER_FULLNAME"]
		mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	} else {
		data["SENDER_FULLNAME"] = data["CLIENT_FULLNAME"]
		mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	}

	var html bytes.Buffer
	if client.DateCreated.After(time.Now().Add(-5 * time.Minute)) {
		newAgreementTpl.ExecuteTemplate(&html, "base", data)
	} else {
		agreementSentTpl.ExecuteTemplate(&html, "base", data)
	}

	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Has Just Sent You A New Agreement"
	mail.Html = html.String()

	_, err := mail.Send()
	return err
}

func AgreementChange(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data, _, _ := createAgreementData(agreement, message, sendToFreelancer)

	mail := new(models.Mail)
	status := agreement.CurrentStatus
	if status.UserID == agreement.FreelancerID {
		data["SENDER_FULLNAME"] = data["FREELANCER_FULLNAME"]
		mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	} else {
		data["SENDER_FULLNAME"] = data["CLIENT_FULLNAME"]
		mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	}

	var html bytes.Buffer
	agreementChangeTpl.ExecuteTemplate(&html, "base", data)

	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Requests Changes to Your Agreement"
	mail.Html = html.String()

	_, err := mail.Send()
	return err
}

func AgreementAccept(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data, _, _ := createAgreementData(agreement, message, sendToFreelancer)
	mail := new(models.Mail)
	status := agreement.CurrentStatus
	if status.UserID == agreement.FreelancerID {
		data["SENDER_FULLNAME"] = data["FREELANCER_FULLNAME"]
		mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	} else {
		data["SENDER_FULLNAME"] = data["CLIENT_FULLNAME"]
		mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	}

	var html bytes.Buffer
	agreementAcceptTpl.ExecuteTemplate(&html, "base", data)

	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Accepted Your Agreement"
	mail.Html = html.String()

	_, err := mail.Send()
	return err
}

func AgreementReject(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	sendToFreelancer := agreement.DraftCreatorID != agreement.FreelancerID
	data, _, _ := createAgreementData(agreement, message, sendToFreelancer)
	mail := new(models.Mail)
	status := agreement.CurrentStatus
	if status.UserID == agreement.FreelancerID {
		data["SENDER_FULLNAME"] = data["FREELANCER_FULLNAME"]
		mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	} else {
		data["SENDER_FULLNAME"] = data["CLIENT_FULLNAME"]
		mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	}

	var html bytes.Buffer
	agreementDisputeTpl.ExecuteTemplate(&html, "base", data)

	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Has Disputed Your Request"
	mail.Html = html.String()

	_, err := mail.Send()
	return err

}

type Agreement struct {
	AgreementID    string     `json:"agreementID"`
	VersionID      string     `json:"versionID" bson:"_id"`
	Version        float64    `json:"version"`
	ClientID       string     `json:"clientID"`
	FreelancerID   string     `json:"freelancerID"`
	Title          string     `json:"title"`
	Payments       []*Payment `json:"payments"`
	WorkItems      WorkItems  `json:"workItems"`
	DraftCreatorID string     `json:"draftCreatorID"`
	CurrentStatus  *Status    `json:"currentStatus"`
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

func (a *Agreement) getTotalCost() float64 {
	var totalCost float64
	workItems := a.WorkItems
	for _, workItem := range workItems {
		totalCost += workItem.Amount
	}
	return totalCost
}

func createAgreementData(agreement *Agreement, message string, toFreelancer bool) (data map[string]interface{}, freelancer *User, client *User) {
	agreementID := agreement.VersionID
	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	path := "/agreement/v/" + agreementID
	client = getUserInfo(clientID)
	freelancer = getUserInfo(freelancerID)
	expiration := 60 * 60 * 24 * 7 * 4
	var signatureParams string
	if toFreelancer {
		signatureParams = createSignatureParams(freelancerID, path, expiration)
	} else {
		signatureParams = createSignatureParams(clientID, path, expiration)
	}

	m := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"CLIENT_FULLNAME":        client.getEmailOrName(),
		"FREELANCER_FULLNAME":    freelancer.getEmailOrName(),
		"MESSAGE":                message,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.WorkItems)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", agreement.getTotalCost()),
		"CLIENT_EMAIL":           client.Email,
		"FREELANCER_EMAIL":       freelancer.Email,
	}
	return m, freelancer, client
}
