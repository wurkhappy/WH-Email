package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"time"
)

type Task struct {
	ID           string    `json:"id"`
	VersionID    string    `json:"versionID"`
	IsPaid       bool      `json:"isPaid"`
	Hours        float64   `json:"hours"`
	SubTasks     []*Task   `json:"subTasks"`
	Title        string    `json:"title"`
	DateExpected time.Time `json:"dateExpected"`
	LastAction   *Action   `json:"lastAction"`
}
type Tasks []*Task

var taskUpdatesTpl = template.Must(template.ParseFiles(
	"templates/base.html",
	"templates/task_updates.html",
))

func TasksUpdated(params map[string]interface{}, body []byte) ([]byte, error, int) {
	fmt.Println(string(body))

	var err error
	var tasks []*Task
	json.Unmarshal(body, &tasks)
	var versionID string
	for _, task := range tasks {
		if task.VersionID != "" {
			versionID = task.VersionID
			break
		}
	}
	var agreement *Agreement
	reply, _ := sendServiceRequest("GET", config.AgreementsService, "/agreements/v/"+versionID, nil)
	json.Unmarshal(reply, &agreement)

	client := getUserInfo(agreement.ClientID)
	freelancer := getUserInfo(agreement.FreelancerID)

	path := "/agreement/v/" + versionID
	expiration := 60 * 60 * 24 * 7 * 4
	signatureParams := createSignatureParams(client.ID, path, expiration, client.IsVerified)
	data := map[string]interface{}{
		"SENDER_FULLNAME": freelancer.createFullName(),
		"Tasks":           tasks,
		"AGREEMENT_LINK":  config.WebServer + path + "?" + signatureParams,
	}
	var html bytes.Buffer
	taskUpdatesTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: client.Email, Name: client.getEmailOrName()}}
	mail.CC = []models.To{{Email: freelancer.Email, Name: freelancer.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = "Status Update For " + agreement.Title
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}
