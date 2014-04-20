package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"strconv"
	"sync"
	"time"
)

var newAgreementTpl *template.Template
var agreementChangeTpl *template.Template
var agreementAcceptTpl *template.Template
var agreementDisputeTpl *template.Template
var agreementRequestTpl *template.Template
var agreementSentTpl *template.Template
var agreementVoidedTpl *template.Template
var agreementSummaryTpl *template.Template

func init() {
	formatDate := func(date time.Time) string {
		return date.Format("Jan 2, 2006")
	}
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
	agreementSummaryTpl, _ = template.New("agreement_summary.html").Funcs(template.FuncMap{"formatDate": formatDate, "unescape": unescaped}).ParseFiles(
		"templates/agreement_summary.html",
	)
}

type agreementAction struct {
	VersionID string `json:"versionID"`
	UserID    string `json:"userID"`
	Message   string `json:"message"`
}

func AgreementSubmitted(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var action *agreementAction
	json.Unmarshal(body, &action)

	agreement, payments, tasks := getAllInfo(action.VersionID)

	if agreement.Version == 1 {
		return newAgreement(agreement, payments, tasks, action.Message)
	}
	return agreementChange(agreement, payments, tasks, action.Message)

}

func newAgreement(agreement *Agreement, payments Payments, tasks Tasks, message string) ([]byte, error, int) {
	var err error
	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.LastAction.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, payments, tasks, message, sender, recipient)

	var html bytes.Buffer
	if !recipient.IsRegistered {
		newAgreementTpl.ExecuteTemplate(&html, "base", data)
	} else {
		agreementSentTpl.ExecuteTemplate(&html, "base", data)
	}

	dataSummary := map[string]interface{}{
		"agreement":  agreement,
		"tasks":      tasks,
		"payments":   payments,
		"freelancer": freelancer,
		"client":     client,
	}
	d, _ := json.Marshal(dataSummary)
	tplSummary, _ := sendServiceRequest("GET", config.PDFTemplatesService, "/template/agreement", d)
	pdfResp, _ := sendServiceRequest("POST", config.PDFService, "/string", tplSummary)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.CC = []models.To{{Email: sender.Email, Name: sender.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Has Sent A New Agreement"
	mail.Html = html.String()
	mail.Attachments = append(mail.Attachments, &models.Attachment{Type: "application/pdf", Name: "Agreement.pdf", Content: string(pdfResp)})

	_, err = mail.Send()
	return nil, err, 200
}

func agreementChange(agreement *Agreement, payments Payments, tasks Tasks, message string) ([]byte, error, int) {
	var err error

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.LastAction.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, payments, tasks, message, sender, recipient)

	var html bytes.Buffer
	agreementChangeTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Requests Changes to Your Agreement"
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}

func AgreementAccept(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var action *agreementAction
	json.Unmarshal(body, &action)
	agreement, payments, tasks := getAllInfo(action.VersionID)

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.LastAction.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, payments, tasks, action.Message, sender, recipient)

	var html bytes.Buffer
	agreementAcceptTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Accepted Your Agreement"
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}

func AgreementReject(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var action *agreementAction
	json.Unmarshal(body, &action)
	agreement, payments, tasks := getAllInfo(action.VersionID)

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	sender, recipient := agreement.LastAction.WhoIsSenderRecipient(client, freelancer)

	data := createAgreementData(agreement, payments, tasks, action.Message, sender, recipient)

	var html bytes.Buffer
	agreementDisputeTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: recipient.Email, Name: recipient.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = data["SENDER_FULLNAME"].(string) + " Has Disputed Your Request"
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}

type Agreement struct {
	AgreementID         string    `json:"agreementID,omitempty"`
	VersionID           string    `json:"versionID,omitempty"` //tracks agreements across versions
	Version             int       `json:"version"`
	ClientID            string    `json:"clientID"`
	FreelancerID        string    `json:"freelancerID"`
	Title               string    `json:"title"`
	ProposedServices    string    `json:"proposedServices"`
	LastModified        time.Time `json:"lastModified"`
	LastAction          *Action   `json:"lastAction"`
	LastSubAction       *Action   `json:"lastSubAction"`
	AcceptsCreditCard   bool      `json:"acceptsCreditCard"`
	AcceptsBankTransfer bool      `json:"acceptsBankTransfer"`
	Archived            bool      `json:"archived"`
}

type Action struct {
	Name   string    `json:"name"`
	Date   time.Time `json:"date"`
	UserID string    `json:"userID"`
	Type   string    `json:"type,omitempty"`
}

func (a *Action) WhoIsSenderRecipient(user1 *User, user2 *User) (sender *User, recipient *User) {
	if a.UserID == user1.ID {
		return user1, user2
	}
	return user2, user1
}

func createAgreementData(agreement *Agreement, payments Payments, tasks []*Task, message string, sender *User, recipient *User) (data map[string]interface{}) {
	agreementID := agreement.VersionID
	path := "/agreement/v/" + agreementID
	expiration := 60 * 60 * 24 * 7 * 4
	signatureParams := createSignatureParams(recipient.ID, path, expiration, recipient.IsVerified)

	m := map[string]interface{}{
		"AGREEMENT_LINK":         config.WebServer + path + "?" + signatureParams,
		"AGREEMENT_NAME":         agreement.Title,
		"SENDER_FULLNAME":        sender.getEmailOrName(),
		"RECIPIENT_FULLNAME":     recipient.getEmailOrName(),
		"MESSAGE":                message,
		"NUM_TASKS":              strconv.Itoa(len(tasks)),
		"AGREEMENT_NUM_PAYMENTS": strconv.Itoa(len(payments)),
		"AGREEMENT_COST":         fmt.Sprintf("%g", payments.getTotalCost()),
	}
	return m
}

func getAllInfo(versionID string) (*Agreement, Payments, Tasks) {
	var agreement *Agreement
	var payments Payments
	var tasks Tasks
	var wg sync.WaitGroup

	wg.Add(3)

	go func(w *sync.WaitGroup) {
		reply, _ := sendServiceRequest("GET", config.AgreementsService, "/agreements/v/"+versionID, nil)
		json.Unmarshal(reply, &agreement)
		w.Done()
	}(&wg)

	go func(w *sync.WaitGroup) {
		reply, _ := sendServiceRequest("GET", config.PaymentsService, "/agreements/v/"+versionID+"/payments", nil)
		json.Unmarshal(reply, &payments)
		w.Done()
	}(&wg)

	go func(w *sync.WaitGroup) {
		reply, _ := sendServiceRequest("GET", config.TasksService, "/agreements/v/"+versionID+"/tasks", nil)
		json.Unmarshal(reply, &tasks)
		w.Done()
	}(&wg)

	wg.Wait()

	return agreement, payments, tasks
}
