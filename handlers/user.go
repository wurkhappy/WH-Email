package handlers

import (
	// "bytes"
	// "encoding/json"
	"github.com/wurkhappy/mandrill-go"
	// "log"
	// "fmt"
	// "net/http"
	// "strconv"
	"time"
)

func ConfirmSignup(params map[string]string, body map[string]interface{}) error {
	userID := body["id"].(string)
	email := body["email"].(string)
	path := "/user/" + userID + "/verify"
	expiration := int(time.Now().Add(time.Hour * 24 * 5).Unix())
	signatureParams := createSignatureParams(userID, path, expiration)
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "signup_link", Content: "http://localhost:4000" + path + "?" + signatureParams},
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

func ForgotPassword(params map[string]string, body map[string]interface{}) error {
	user := body["user"].(map[string]interface{})
	userID := user["id"].(string)
	email := user["email"].(string)
	path := "/user/new-password"
	expiration := int(time.Now().Add(time.Hour * 1).Unix())
	signatureParams := createSignatureParams(userID, path, expiration)
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "PASSWORD_RESET_LINK", Content: "http://localhost:4000" + path + "?" + signatureParams},
		&mandrill.GlobalVar{Name: "USER_FULLNAME", Content: createFullName(user)},
	)
	message.To = []mandrill.To{{Email: email, Name: createFullName(user)}}
	m.Args["message"] = message
	m.Args["template_name"] = "User Reset Password"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}
