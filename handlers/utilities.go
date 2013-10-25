package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/http"
	"time"
)

func sendRequest(r *http.Request) (map[string]interface{}, []byte) {
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
	respBuf := new(bytes.Buffer)
	respBuf.ReadFrom(resp.Body)
	var respData map[string]interface{}
	json.Unmarshal(respBuf.Bytes(), &respData)
	return respData, respBuf.Bytes()
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

func signURL(userID, path, method string, expiration int) string {
	tkn, _ := uuid.NewV4()
	token := tkn.String()
	expirationDate := int(time.Now().Add(time.Duration(expiration) * time.Second).Unix())
	c := redisPool.Get()
	if _, err := c.Do("HMSET", token, "path", path, "method", method, "expiration", expirationDate, "userID", userID); err != nil {
		log.Panic(err)
	}

	if _, err := c.Do("EXPIRE", token, expiration); err != nil {
		log.Panic(err)
	}

	return token
}
func createSignatureParams(userID, path string, expiration int) string {
	token := signURL(userID, path, "GET", expiration)
	return "token=" + token
}

func getUserMessage(body map[string]interface{}) string {
	var userMessage string = " "
	if msg, ok := body["message"]; ok {
		userMessage = msg.(string)
	}

	return userMessage
}

func getEmailOrName(user map[string]interface{}) string {
	name := createFullName(user)
	if name == "" {
		name = user["email"].(string)
	}

	return name
}



func Test(params map[string]string, body map[string]interface{}) error {
	time.Sleep(time.Second * 1)
	log.Print(body)
	return nil
}
