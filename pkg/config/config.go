package config

import (
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/repository"
	"log"
	"text/template"

	"go.mongodb.org/mongo-driver/mongo"
)

type AppWideConfig struct {
	Properties    map[string]string
	UseCache      bool
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	MailChannel   chan models.MailData
	TemplateCache map[string]*template.Template
	MongoClient   *mongo.Client
	AlertRepo     *repository.AlertRepository
}
