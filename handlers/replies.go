package handlers

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/wurkhappy/WH-Config"
	"fmt"
)

type Reply struct {
	HTML        []string `json:"stripped-html"`
	InReplyToID []string `json:"In-Reply-To"`
}

func ProcessReply(params map[string]string, body map[string]*json.RawMessage) error {
	var emailBytes []byte
	json.Unmarshal(*body["message"], &emailBytes)
	//var rep map[string]interface{}
	reply := new(Reply)
	err := json.Unmarshal(emailBytes, &reply)
	fmt.Println(reply)
	fmt.Println(err)
	fmt.Println(reply.InReplyToID[0])
	c := redisPool.Get()
	commentBytes, err := redis.Bytes(c.Do("GET", reply.InReplyToID[0]))
	fmt.Println("redis", string(commentBytes), err)
	var comment *Comment
	json.Unmarshal(commentBytes, &comment)
	newComment := new(Comment)
	newComment.AgreementID = comment.AgreementID
	newComment.Tags = comment.Tags
	newComment.Text = reply.HTML[0]
	newComment.UserID = comment.RecipientID

	newCommentjson, _ := json.Marshal(newComment)
	_, statusCode := sendServiceRequest("POST", config.CommentsService, "/agreement/"+comment.AgreementID+"/comments", newCommentjson)
	if statusCode >= 400 {
		return nil
	}
	return nil
}
