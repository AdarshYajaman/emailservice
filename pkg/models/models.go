package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MailData holds a message
type MailData struct {
	To       []string
	From     string
	Subject  string
	Content  map[string]interface{}
	Template string
}

// Alert holds fields that are required to process incoming request to email service
type Alert struct {
	IndexId       primitive.ObjectID `bson:"_id"`
	MigrationId   string             `json:"migrationId"`
	Volumes       []string           `json:"volumes"`
	AlertType     string             `json:"alertType"`
	MigrationDate time.Time          `json:"migrationDate"`
	// AlertSchedule string
	// TemplateName  string
	AlertStatus   bool
	AlertSentTime time.Time
	IsReadyToSend bool
}

// Job holds schedule details to run
type Job struct {
	IndexId        primitive.ObjectID `bson:"_id"`
	CronExpression string
	Comments       string
	FromDay        uint8
	ToDay          uint8
	TemplateName   string
	AddedBy        string
	CreatedAt      time.Time
	MailSubject    string
}
