package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

func getUserInfo(id string) map[string]interface{} {
	if id == "" {
		return make(map[string]interface{})
	}
	client := &http.Client{}
	r, _ := http.NewRequest("GET", "http://localhost:3000/user/search?userid="+id, nil)
	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
	clientBuf := new(bytes.Buffer)
	clientBuf.ReadFrom(resp.Body)
	var clientData []map[string]interface{}
	json.Unmarshal(clientBuf.Bytes(), &clientData)
	if len(clientData) > 0 {
		return clientData[0]
	}
	return nil
}

func signURL(userID, path, expiration string) string {
	body := bytes.NewReader([]byte(`{"path":"` + path + `", "expiration":` + expiration + `}`))
	r, _ := http.NewRequest("POST", "http://localhost:3000/user/"+userID+"/sign", body)
	respData, _ := sendRequest(r)
	signature := respData["signature"].(string)
	return signature
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

func getTotalCost(agreement map[string]interface{}) float64 {
	var totalCost float64
	payments := agreement["payments"].([]interface{})
	for _, payment := range payments {
		model := payment.(map[string]interface{})
		totalCost += model["amount"].(float64)
	}

	return totalCost

}

func createSignatureParams(userID, path string, expiration int) string {
	exp := strconv.Itoa(expiration)
	signature := signURL(userID, path, exp)
	return "signature=" + signature + "&access_key=" + userID + "&expiration=" + exp
}

func Test(params map[string]string, body map[string]interface{}) error {
	time.Sleep(time.Second * 1)
	log.Print(body)
	return nil
}
