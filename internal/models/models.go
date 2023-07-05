package models

import (
	"github.com/minhnguyen/internal/forms"
)

type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	Data      map[string]interface{}
	CSRFToken string
	Flash     string
	Warning   string
	Error     string
	Form      *forms.Form
}

type MailData struct {
	To string
	From string
	Subject string
	Content string
	Template string
}