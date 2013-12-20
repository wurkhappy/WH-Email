package handlers

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/wurkhappy/WH-Config"
	"io"
	"net/url"
)

func ProcessReply(params map[string]string, body map[string]*json.RawMessage) error {
	var str string
	json.Unmarshal(*body["message"], &str)
	fmt.Println(str)
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
