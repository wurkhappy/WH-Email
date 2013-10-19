package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
