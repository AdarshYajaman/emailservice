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
	IndexId          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	MigrationId      string             `json:"migrationId,omitempty" bson:"migrationId,omitempty"`
	Volumes          []string           `json:"volumes,omitempty" bson:"volumes,omitempty"`
	AlertType        string             `json:"alertType,omitempty" bson:"alertType,omitempty"`
	MigrationDate    time.Time          `json:"migrationDate,omitempty" bson:"migrationDate,omitempty"`
	DistributionList []string           `json:"distributionList,omitempty" bson:"distributionList,omitempty"`
	AlertStatus      string             `json:"alertStatus,omitempty" bson:"alertStatus,omitempty"`
	AlertSentTime    time.Time          `json:"alertSentTime,omitempty" bson:"alertSentTime,omitempty"`
	IsReadyToSend    *bool              `json:"isReadyToSend,omitempty" bson:"isReadyToSend,omitempty"`
}

// Job holds schedule details to run
type Job struct {
	IndexId        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CronExpression string             `json:"cronExpression,omitempty" bson:"cronExpression,omitempty"`
	Comments       string             `json:"comments,omitempty" bson:"comments,omitempty"`
	StartDate      uint8              `json:"startDate,omitempty" bson:"startDate,omitempty"`
	EndDate        uint8              `json:"endDate,omitempty" bson:"endDate,omitempty"`
	TemplateName   string             `json:"templateName,omitempty" bson:"templateName,omitempty"`
	AddedBy        string             `json:"addedBy,omitempty" bson:"addedBy,omitempty"`
	CreatedAt      time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	MailSubject    string             `json:"mailSubject,omitempty" bson:"mailSubject,omitempty"`
}

// ErrorResponse holds error details
type ErrorResponse struct {
	ErrorMessage string
}

type NotifyMigrationDateChange struct {
	IsReadyToSend    *bool     `json:"isReadyToSend"`
	NewMigrationDate time.Time `json:"newMigrationDate,omitempty"`
}
