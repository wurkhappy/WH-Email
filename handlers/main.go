package handlers

import (
	"html/template"
)

var templates *template.Template

func init() {
	templates = template.Must(template.ParseGlob("templates/*"))
}
