package handlers

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
		"github.com/garyburd/redigo/redis"
)

func ProcessReply(params map[string]string, body map[string]*json.RawMessage) error {
	var str string
	json.Unmarshal(*body["message"], &str)
	s, _ := url.QueryUnescape(str)
	s = s[16:]
	var m []map[string]interface{}
	json.Unmarshal([]byte(s), &m)
	to := m[0]["msg"].(map[string]interface{})["to"].([]interface{})
	t := to[0]
	toString := t.([]interface{})[0].(string)
	message_id := toString[6:42]
	fmt.Println(message_id)
	msg := m[0]["msg"].(map[string]interface{})["html"]
	r := bytes.NewReader([]byte(msg.(string)))
	comment := parseHtml(r)
	fmt.Println(comment)

	commentBytes := redis.Bytes(c.Send("GET", message_id))
	fmt.Println(string(commentBytes))
	return nil
}

func parseHtml(r io.Reader) string {
	depth := 0
	var s string
	d := html.NewTokenizer(r)
	for {
		tokenType := d.Next()
		if tokenType == html.ErrorToken {
			return ""
		}
		token := d.Token()
		switch tokenType {
		case html.StartTagToken:
			if token.Data != "br" {
				depth += 1
			}
			s += token.String()
		case html.TextToken:
			s += token.String()
		case html.EndTagToken:
			depth -= 1
			s += token.String()
			if depth == 0 {
				return s
			}
		case html.SelfClosingTagToken:

		}
	}
	return ""
}
