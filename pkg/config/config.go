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
	DefaultAlert        string `mapstructure:"default_alert"`
	MongoURL            string `mapstructure:"mongo_url"`
	MongoDBName         string `mapstructure:"mongo_dbname"`
	MongoCollectionName string `mapstructure:"mongo_collectionname"`
	MongoTimeout        string `mapstructure:"mongo_timeout"`
	DefaultTemplate     string `mapstructure:"default_template"`
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
	AlertRepo         *repository.AlertRepository
	CronJobs          *cron.Cron
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
