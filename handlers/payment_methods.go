package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
)

var BankAccountMissingTpl = template.Must(template.ParseFiles(
	"templates/base.html",
	"templates/add_bank.html",
))

var BankAccountVerifyTpl = template.Must(template.ParseFiles(
	"templates/base.html",
	"templates/new_bank.html",
))

func BankAccountMissing(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var message struct {
		UserID string `json:"userID"`
	}
	json.Unmarshal(body, &message)
	user := getUserInfo(message.UserID)

	path := "/account"
	expiration := 60 * 60 * 24 * 7 * 4
	signatureParams := createSignatureParams(user.ID, path, expiration, user.IsVerified)

	data := map[string]interface{}{
		"RECIPIENT_FULLNAME":      user.createFullName(),
		"BANK_ACCOUNT_SETUP_LINK": config.WebServer + path + "?" + signatureParams + "#bankaccount",
	}
	var html bytes.Buffer
	BankAccountMissingTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: user.Email, Name: user.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = "Bank Account Required"
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}

func BankAccountAdded(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var err error
	var message struct {
		UserID string `json:"userID"`
	}
	json.Unmarshal(body, &message)
	user := getUserInfo(message.UserID)

	path := "/account"
	expiration := 60 * 60 * 24 * 7 * 4
	signatureParams := createSignatureParams(user.ID, path, expiration, user.IsVerified)

	data := map[string]interface{}{
		"RECIPIENT_FULLNAME":      user.createFullName(),
		"BANK_ACCOUNT_SETUP_LINK": config.WebServer + path + "?" + signatureParams + "#bankaccount",
	}
	var html bytes.Buffer
	BankAccountVerifyTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: user.Email, Name: user.getEmailOrName()}}
	mail.FromEmail = "info@notifications.wurkhappy.com"
	mail.Subject = "New Bank Account Added! Please Verify."
	mail.Html = html.String()

	_, err = mail.Send()
	return nil, err, 200
}
