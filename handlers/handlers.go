package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/wurkhappy/mandrill-go"
	"log"
	"net/http"
)

func init() {
	mandrill.APIkey = "tKcqIfanhMnYrTtGrDixBA"
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
		&mandrill.GlobalVar{Name: "signup_link", Content: "http://localhost:4000" + path + "?signature=" + signature},
	)
	message.To = []mandrill.To{{Email: email}}
	m.Args["message"] = message
	m.Args["template_name"] = "Confirm Email and Signup"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func AgreementToNewUser(params map[string]string, body map[string]interface{}) error {
	userID := body["id"].(string)
	email := body["email"].(string)
	path := "/user/" + userID + "/verify"
	signature := signURL(userID, path)
	
	m := mandrill.NewCall()
	m.Category = "messages"
	m.Method = "send-template"
	message := new(mandrill.Message)
	message.GlobalMergeVars = append(message.GlobalMergeVars,
		&mandrill.GlobalVar{Name: "signup_link", Content: "http://localhost:4000" + path + "?signature=" + signature},
	)
	message.To = []mandrill.To{{Email: email}}
	m.Args["message"] = message
	m.Args["template_name"] = "Confirm Email and Signup"
	m.Args["template_content"] = []mandrill.TemplateContent{{Name: "blah", Content: "nfd;jd;fjvnbd"}}

	_, err := m.Send()
	if err != nil {
		return err
	}
	return nil
}

func signURL(userID, path string) string {
	body := bytes.NewReader([]byte(`{"path":"` + path + `"}`))
	r, _ := http.NewRequest("POST", "http://localhost:3000/user/"+userID+"/sign", body)
	respData, _ := sendRequest(r)
	signature := respData["signature"].(string)
	return signature
}
