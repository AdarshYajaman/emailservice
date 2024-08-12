package config

import (
	"log"
	"text/template"

	"citi.com/179563_genesis_mail/pkg/models"
	"citi.com/179563_genesis_mail/pkg/repository"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
)

type ApplicationProperties struct {
	APIPort             string `mapstructure:"api_port"`
	APITimeOut          int    `mapstructure:"api_timeOutInSeconds"`
	SMTPHost            string `mapstructure:"smtp_host"`
	SMTPPort            string `mapstructure:"smtp_port"`
	SMTPTimeOut         string `mapstructure:"smtp_timeOutInSeconds"`
	SMTPChannelBufSize  int    `mapstructure:"smtp_channelbufsize"`
	SMTPWorkers         int    `mapstructure:"smtp_workers"`
	DefaultJobRefresh   string `mapstructure:"default_jobrefresh"`
	DefaultTemplate     string `mapstructure:"default_template"`
	FromAddress         string `mapstructure:"default_fromaddress"`
	MongoURL            string `mapstructure:"mongo_url"`
	MongoDBName         string `mapstructure:"mongo_dbname"`
	MongoTimeout        int    `mapstructure:"mongo_timeOutInSeconds"`
	AlertCollectionName string `mapstructure:"mongo_alertcollectionname"`
	JobCollectionName   string `mapstructure:"mongo_jobcollectionname"`
}

// func (props *ApplicationProperties) String() string {
// 	return fmt.Sprintf("props.SMTPHost - %s", props.SMTPHost)
// }

type AppWideConfig struct {
	Properties        *ApplicationProperties
	UseCache          bool
	InfoLog           *log.Logger
	ErrorLog          *log.Logger
	MailChannel       chan models.MailData
	MailTemplateCache map[string]*template.Template
	MongoClient       *mongo.Client
	// AlertRepo         repository.AlertRepository
	// JobRepo           repository.JobRepository

	AlertRepo repository.Repository[*models.Alert]
	JobRepo   repository.Repository[*models.Job]

	CronJobs *cron.Cron
	JobMap   map[string]*models.Job
}
