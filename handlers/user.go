package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/WH-Email/models"
	"html/template"
	"time"
	//"fmt"
)

var confirmSignupTpl *template.Template
var passwordResetTpl *template.Template

func init() {
	confirmSignupTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/confirm_signup.html",
	))
	passwordResetTpl = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/password_reset.html",
	))
}

func ConfirmSignup(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var user *User
	json.Unmarshal(body, &user)
	userID := user.ID
	email := user.Email

	path := "/user/" + userID + "/verify"
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(userID, path, expiration, user.IsVerified)

	data := map[string]interface{}{
		"SIGNUP_LINK": config.WebServer + path + "?" + signatureParams,
	}
	var html bytes.Buffer
	confirmSignupTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: email, Name: user.createFullName()}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = "Welcome to Wurk Happy!"
	mail.Html = html.String()

	_, err := mail.Send()
	return nil, err, 200
}

func ForgotPassword(params map[string]interface{}, body []byte) ([]byte, error, int) {
	var user *User
	json.Unmarshal(body, &user)
	userID := user.ID
	email := user.Email

	path := "/user/new-password"
	expiration := int(time.Now().Add(time.Hour * 1).Unix())
	signatureParams := createSignatureParams(userID, path, expiration, user.IsVerified)

	data := map[string]interface{}{
		"PASSWORD_RESET_LINK": config.WebServer + path + "?" + signatureParams,
		"USER_FULLNAME":       user.createFullName(),
	}
	var html bytes.Buffer
	passwordResetTpl.ExecuteTemplate(&html, "base", data)

	mail := new(models.Mail)
	mail.To = []models.To{{Email: email, Name: user.createFullName()}}
	mail.FromEmail = "reply@notifications.wurkhappy.com"
	mail.FromName = "Wurk Happy"
	mail.Subject = "Wurk Happy Request to Reset Password"
	mail.Html = html.String()

	_, err := mail.Send()
	return nil, err, 200
}
