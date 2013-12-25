package handlers

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/wurkhappy/WH-Config"
)

type Reply struct {
	HTML        []string `json:"stripped-html"`
	InReplyToID []string `json:"In-Reply-To"`
}

func ProcessReply(params map[string]string, body map[string]*json.RawMessage) error {
	var emailBytes []byte
	json.Unmarshal(*body["message"], &emailBytes)
	reply := new(Reply)
	json.Unmarshal(emailBytes, &reply)
	c := redisPool.Get()
	commentBytes, _ := redis.Bytes(c.Do("GET", reply.InReplyToID[0]))
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
