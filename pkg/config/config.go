package config

import (
	"103-EmailService/pkg/models"
	"log"
	"text/template"
)

type AppWideConfig struct {
	Properties    map[string]string
	UseCache      bool
	Logger        *log.Logger
	MailChannel   chan models.MailData
	TemplateCache map[string]*template.Template
}
