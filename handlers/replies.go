package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/wurkhappy/WH-Config"
	"strings"
)

type Reply struct {
	HTML       []string `json:"stripped-html"`
	References []string `json:"References"`
	From       []string `json:"From"`
}

func ProcessReply(params map[string]string, body map[string]*json.RawMessage) error {
	var emailBytes []byte
	json.Unmarshal(*body["message"], &emailBytes)
	reply := new(Reply)
	json.Unmarshal(emailBytes, &reply)
	refs := reply.References[0]
	index := strings.Index(refs, ">")
	msgID := refs[0 : index+1]

	var messageInfo struct {
		Comment    string `redis:"comment"`
		User1Email string `redis:"user1Email"`
		User1ID    string `redis:"user1ID"`
		User2Email string `redis:"user2Email"`
		User2ID    string `redis:"user2ID"`
	}

	c := redisPool.Get()
	v, _ := redis.Values(c.Do("HGETALL", msgID))
	if err := redis.ScanStruct(v, &messageInfo); err != nil {
		fmt.Printf("%s", "There was an error parsing that token")
	}
	var comment *Comment
	json.Unmarshal([]byte(messageInfo.Comment), &comment)
	newComment := new(Comment)
	newComment.AgreementID = comment.AgreementID
	newComment.Tags = comment.Tags
	newComment.Text = reply.HTML[0]

	senderID := messageInfo.User1ID
	if strings.Index(reply.From[0], messageInfo.User1Email) == -1 {
		senderID = messageInfo.User2ID
	}
	newComment.UserID = senderID

	newCommentjson, _ := json.Marshal(newComment)
	_, statusCode := sendServiceRequest("POST", config.CommentsService, "/agreement/"+comment.AgreementID+"/comments", newCommentjson)
	if statusCode >= 400 {
		return nil
	}
	return nil
}
