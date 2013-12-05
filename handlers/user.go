package handlers

import (
	"encoding/json"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/mandrill-go"
	"time"
	"fmt"
)

func ConfirmSignup(params map[string]string, body map[string]*json.RawMessage) error {
	var user *User
	json.Unmarshal(*body["user"], &user)
	userID := user.ID
	email := user.Email

	path := "/user/" + userID + "/verify"
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(userID, path, expiration)
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "signup_link", Content: config.WebServer + path + "?" + signatureParams},
	)
	message.To = []mandrill.To{{Email: email, Name: user.createFullName()}}
	message.Tags = []string{"test", "userid"}
	mp := map[string]string{
		"userid": "123",
	}
	message.Metadata = []map[string]string{mp}
	m.Args["message"] = message
	m.Args["template_name"] = "Confirm Email and Signup"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return fmt.Errorf("%s", err.Message)
	}
	return nil
}

func ForgotPassword(params map[string]string, body map[string]*json.RawMessage) error {
	var user *User
	json.Unmarshal(*body["user"], &user)
	userID := user.ID
	email := user.Email

	path := "/user/new-password"
	expiration := int(time.Now().Add(time.Hour * 1).Unix())
	signatureParams := createSignatureParams(userID, path, expiration)
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "PASSWORD_RESET_LINK", Content: config.WebServer + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "USER_FULLNAME", Content: user.createFullName()},
	)
	message.To = []mandrill.To{{Email: email, Name: user.createFullName()}}
	m.Args["message"] = message
	m.Args["template_name"] = "User Reset Password"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return fmt.Errorf("%s", err.Message)
	}
	return nil
}
