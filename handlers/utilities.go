package handlers

import (
	"encoding/json"
	"github.com/nu7hatch/gouuid"
	"github.com/wurkhappy/WH-Config"
	"github.com/wurkhappy/mdp"
	"html/template"
	"log"
	"time"
)

type ServiceResp struct {
	StatusCode float64 `json:"status_code"`
	Body       []byte  `json:"body"`
}

func unescaped(x string) interface{} { return template.HTML(x) }

func sendServiceRequest(method, service, path string, body []byte) (response []byte, statusCode int) {
	client := mdp.NewClient(config.MDPBroker, false)
	client.Timeout = 10 * time.Second
	defer client.Close()
	m := map[string]interface{}{
		"Method": method,
		"Path":   path,
		"Body":   body,
	}
	req, _ := json.Marshal(m)
	request := [][]byte{req}
	reply := client.Send([]byte(service), request)
	if len(reply) == 0 {
		return nil, 404
	}
	resp := new(ServiceResp)
	json.Unmarshal(reply[0], &resp)
	return resp.Body, int(resp.StatusCode)
}

func createFullName(user map[string]interface{}) string {
	fName, fnOK := user["firstName"]
	lName, lnOK := user["lastName"]
	var fullname string
	if fnOK && lnOK {
		fullname = fName.(string) + " " + lName.(string)
	} else if fnOK {
		fullname = fName.(string)
	}

	return fullname
}

func signURL(userID, path, method string, expiration int, verified bool) string {
	tkn, _ := uuid.NewV4()
	token := tkn.String()
	expirationDate := int(time.Now().Add(time.Duration(expiration) * time.Second).Unix())
	c := redisPool.Get()
	if _, err := c.Do("HMSET", token, "path", path, "method", method, "expiration", expirationDate, "userID", userID, "verified", verified); err != nil {
		log.Panic(err)
	}

	if _, err := c.Do("EXPIRE", token, expiration); err != nil {
		log.Panic(err)
	}

	return token
}
func createSignatureParams(userID, path string, expiration int, verified bool) string {
	token := signURL(userID, path, "GET", expiration, verified)
	return "token=" + token
}

func getEmailOrName(user map[string]interface{}) string {
	name := createFullName(user)
	if name == "" {
		name = user["email"].(string)
	}

	return name
}

func getAgreementOwners(agreementID string) *Agreement {
	resp, statusCode := sendServiceRequest("GET", config.AgreementsService, "/agreements/"+agreementID+"/owners", nil)
	if statusCode >= 400 {
		return nil
	}
	var a *Agreement
	json.Unmarshal(resp, &a)
	return a
}
