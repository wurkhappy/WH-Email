package handlers

import (
	"bytes"
	"encoding/json"
	// "github.com/wurkhappy/mandrill-go"
	"fmt"
	// "log"
	"net/http"
	// "strconv"
	// "time"
)

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (u *User) getEmailOrName() string {
	name := u.createFullName()
	if name == "" {
		name = u.Email
	}

	return name
}

func getUserInfo(id string) *User {
	if id == "" {
		return nil
	}
	client := &http.Client{}
	r, _ := http.NewRequest("GET", UserService+"/user/search?userid="+id, nil)
	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
	clientBuf := new(bytes.Buffer)
	clientBuf.ReadFrom(resp.Body)
	var users []*User
	json.Unmarshal(clientBuf.Bytes(), &users)
	if len(users) > 0 {
		return users[0]
	}
	return nil
}

func (u *User) createFullName() string {
	fName := u.FirstName
	lName := u.LastName
	fnOK := fName != ""
	lnOK := lName != ""

	var fullname string
	if fnOK && lnOK {
		fullname = fName + " " + lName
	} else if fnOK {
		fullname = fName
	}

	return fullname
}
