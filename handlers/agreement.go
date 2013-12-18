package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"strconv"
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
		"templates/agreement_accept.html",
	))
	agreementRequestTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_accept.html",
	))
	agreementSentTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_accept.html",
	))
	agreementVoidedTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/agreement_accept.html",
	))
}

func NewAgreement(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	data := createAgreementData(agreement, message)

	var html bytes.Buffer
	newAgreementTpl.Execute(&html, data)

	mail := new(models.Mail)
	if agreement.DraftCreatorID == agreement.FreelancerID {
		mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	} else {
		mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["FREELANCER_FULLNAME"].(string) + " Has Just Sent You A New Agreement"
	mail.Html = html.String()

	return mail.Send()
}

func AgreementChange(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	data := createAgreementData(agreement, message)

	var html bytes.Buffer
	agreementChangeTpl.Execute(&html, data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["CLIENT_EMAIL"].(string), Name: data["CLIENT_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["FREELANCER_FULLNAME"].(string) + " Requests Changes to Your Agreement"
	mail.Html = html.String()

	return mail.Send()
}

func AgreementAccept(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	data := createAgreementData(agreement, message)

	var html bytes.Buffer
	agreementAcceptTpl.Execute(&html, data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["CLIENT_FULLNAME"].(string) + " Accepted Your Agreement"
	mail.Html = html.String()

	return mail.Send()
}

func AgreementReject(params map[string]string, body map[string]*json.RawMessage) error {
	var agreement *Agreement
	json.Unmarshal(*body["agreement"], &agreement)
	var message string
	if messageBytes, ok := body["message"]; ok {
		json.Unmarshal(*messageBytes, &message)
	}

	data := createAgreementData(agreement, message)

	var html bytes.Buffer
	agreementDisputeTpl.Execute(&html, data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: data["FREELANCER_EMAIL"].(string), Name: data["FREELANCER_FULLNAME"].(string)}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = data["CLIENT_FULLNAME"].(string) + " Has Disputed Your Request"
	mail.Html = html.String()

	return mail.Send()

}

type Agreement struct {
	AgreementID    string     `json:"agreementID"`
	VersionID      string     `json:"versionID" bson:"_id"`
	Version        float64    `json:"version"`
	ClientID       string     `json:"clientID"`
	FreelancerID   string     `json:"freelancerID"`
	Title          string     `json:"title"`
	Payments       []*Payment `json:"payments"`
	DraftCreatorID string     `json:"draftCreatorID"`
}

func (a *Agreement) getTotalCost() float64 {
	var totalCost float64
	payments := a.Payments
	for _, payment := range payments {
		totalCost += payment.Amount
	}
	return totalCost
}

func createAgreementData(agreement *Agreement, message string) map[string]interface{} {
	agreementID := agreement.VersionID
	clientID := agreement.ClientID
	freelancerID := agreement.FreelancerID
	path := "/agreement/v/" + agreementID
	client := getUserInfo(clientID)
	freelancer := getUserInfo(freelancerID)
	expiration := 60 * 60 * 24
	signatureParams := createSignatureParams(freelancerID, path, expiration)

	m := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"CLIENT_FULLNAME":        client.getEmailOrName(),
		"FREELANCER_FULLNAME":    freelancer.getEmailOrName(),
		"MESSAGE":                message,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.Payments)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", agreement.getTotalCost()),
		"CLIENT_EMAIL":           client.Email,
		"FREELANCER_EMAIL":       freelancer.Email,
	}
	return m
}
