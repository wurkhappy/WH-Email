package handlers

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/wurkhappy/WH-Config"
	//"fmt"
	"strings"
)

type Reply struct {
	HTML        []string `json:"stripped-html"`
	References []string `json:"References"`
}

func ProcessReply(params map[string]string, body map[string]*json.RawMessage) error {
	var emailBytes []byte
	json.Unmarshal(*body["message"], &emailBytes)
	reply := new(Reply)
	json.Unmarshal(emailBytes, &reply)
	refs := reply.References[0]
	index := strings.Index(refs, ">")
	msgID := refs[0:index+1]
	
	c := redisPool.Get()
	commentBytes, _ := redis.String(c.Do("GET", msgID))
	var comment *Comment
	json.Unmarshal([]byte(commentBytes), &comment)
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
