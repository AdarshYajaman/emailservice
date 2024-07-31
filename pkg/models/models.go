package models

import (
	"time"
)

// MailData holds a message
type MailData struct {
	To       []string
	From     string
	Subject  string
	Content  map[string]string
	Template string
}

// Alert holds fields that are required to process incoming request to email service
type Alert struct {
	MigrationId   string    `json:"id"`
	Volumes       []string  `json:"volumes"`
	AlertType     string    `json:"alertType"`
	MigrationDate time.Time `json:"migrationDate"`
	AlertSchedule string
	TemplateName  string
	AlertStatus   bool
	AlertTime     time.Time
}
