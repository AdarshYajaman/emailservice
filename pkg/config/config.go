package config

import (
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/repository"
	"fmt"
	"log"
	"text/template"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

type ApplicationProperties struct {
	SMTPHost            string `mapstructure:"smtp_host"`
	SMTPPort            string `mapstructure:"smtp_port"`
	SMTPTimeOut         string `mapstructure:"smtp_timeout"`
	SMTPChannelBufSize  int    `mapstructure:"smtp_channelbufsize"`
	SMTPWorkers         int    `mapstructure:"smtp_workers"`
	DefaultJobRefresh   string `mapstructure:"default_jobrefresh"`
	DefaultTemplate     string `mapstructure:"default_template"`
	FromAddress         string `mapstructure:"default_fromaddress"`
	MongoURL            string `mapstructure:"mongo_url"`
	MongoDBName         string `mapstructure:"mongo_dbname"`
	MongoTimeout        string `mapstructure:"mongo_timeout"`
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
	AlertRepo         repository.AlertRepository
	JobRepo           repository.JobRepository

	// AlertRepo         repository.Repository[*models.Alert]
	// JobRepo           repository.Repository[*models.Job]

	CronJobs *cron.Cron
	JobMap   map[string]*models.Job
}

func ReadConfigFile() *ApplicationProperties {
	appProps := ApplicationProperties{}

	viper.SetConfigName("application")
	viper.AddConfigPath("./pkg/config/")
	viper.SetConfigType("properties")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	err = viper.Unmarshal(&appProps)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	return &appProps

}
