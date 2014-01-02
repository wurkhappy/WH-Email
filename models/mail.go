package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

var production bool

func Setup(production bool) {
	production = production
}

type Mail struct {
	Html        string        `json:"html,omitempty"`
	Subject     string        `json:"subject"`
	FromEmail   string        `json:"from_email"`
	FromName    string        `json:"from_name"`
	To          []To          `json:"to"`
	BCCAddress  string        `json:"bcc_address,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type To struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type MailGunResp struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (mail *Mail) Send() (msgID string, erro error) {
	var err error
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	err = w.WriteField("from", mail.FromName+" <"+mail.FromEmail+">")
	if err != nil {
		return "", err
	}

	for _, recipient := range mail.To {
		err = w.WriteField("to", recipient.Name+" <"+recipient.Email+">")
		if err != nil {
			return "", err
		}
	}

	err = w.WriteField("subject", mail.Subject)
	if err != nil {
		return "", err
	}

	err = w.WriteField("html", mail.Html)
	if err != nil {
		return "", err
	}
	// if !production {
	// 	err = w.WriteField("o:testmode", "true")
	// 	if err != nil {
	// 		return "", err
	// 	}
	// }

	for _, attachment := range mail.Attachments {
		attach, err := w.CreateFormFile("attachment", attachment.Name)
		if err != nil {
			return "", err
		}
		attach.Write([]byte(attachment.Content))
	}
	w.Close()
	req, err := http.NewRequest("POST", "https://api.mailgun.net/v2/notifications.wurkhappy.com/messages", buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth("api", "key-6t8u9z7c059n0is1c6k4779flnen1zf3")
	res, err := http.DefaultClient.Do(req)
	resbuf := new(bytes.Buffer)
	resbuf.ReadFrom(res.Body)
	mgResp := new(MailGunResp)
	fmt.Println(resbuf.String())
	json.Unmarshal(resbuf.Bytes(), &mgResp)
	if mgResp.ID == "" {
		err = fmt.Errorf(mgResp.Message)
	}
	return mgResp.ID, err

}
