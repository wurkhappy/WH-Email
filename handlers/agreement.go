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
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	if client.DateCreated.After(time.Now().Add(-5 * time.Minute)) {
		data["AGREEMENT_LINK"] = data["AGREEMENT_LINK"].(string) + "#new-account"
		newAgreementTpl.ExecuteTemplate(&html, "base", data)
	} else {
		agreementSentTpl.ExecuteTemplate(&html, "base", data)
	}

	mail := new(models.Mail)
	mail.To = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
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
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	agreementChangeTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
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
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	agreementAcceptTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
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
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.CurrentStatus.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, message, sender, recipient)

	var html bytes.Buffer
	agreementDisputeTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
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

func (s *Status) WhoIsSenderRecipient(user1 *User, user2 *User) (sender *User, recipient *User) {
	if s.UserID == user1.ID {
		return user1, user2
	}
	return user2, user1
}

func (a *Agreement) getTotalCost() float64 {
	var totalCost float64
	workItems := a.WorkItems
	for _, workItem := range workItems {
		totalCost += workItem.Amount
	}
	return totalCost
}

func createAgreementData(agreement *Agreement, message string, sender *User, recipient *User) (data map[string]interface{}) {
	agreementID := agreement.VersionID
	path := "/agreement/v/" + agreementID
	expiration := 60 * 60 * 24 * 7 * 4
	signatureParams := createSignatureParams(recipient.ID, path, expiration)

	m := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"SENDER_FULLNAME":        sender.getEmailOrName(),
		"RECIPIENT_FULLNAME":     recipient.getEmailOrName(),
		"MESSAGE":                message,
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(agreement.WorkItems)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", agreement.getTotalCost()),
	}
	return m
}
