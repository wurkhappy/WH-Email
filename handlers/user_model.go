package handlers

import (
	"encoding/json"
	"github.com/wurkhappy/WH-Config"
	"time"
)

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	DateCreated time.Time `json:"dateCreated"`
	IsVerified  bool      `json:"isVerified"`
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
	resp, statusCode := sendServiceRequest("GET", config.UserService, "/user/search?userid="+id, nil)
	if statusCode >= 400 {
		return nil
	}
	var users []*User
	json.Unmarshal(resp, &users)
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
