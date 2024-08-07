package main

import (
	"103-EmailService/pkg/config"
	"103-EmailService/pkg/handler"
	"103-EmailService/pkg/helpers"
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/repository"
	"103-EmailService/pkg/scheduler"
	"103-EmailService/pkg/service"
	"context"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var appConfig config.AppWideConfig

func main() {

	configure()

	defer close(appConfig.MailChannel)

	defer func() {
		if err := appConfig.MongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	router := http.NewServeMux()

	router.HandleFunc("GET /api/alerts", handler.GetAlerts)
	router.HandleFunc("POST /api/alert", handler.CreateAlert)
	router.HandleFunc("PATCH /api/alert", handler.UpdateAlert)
	router.HandleFunc("DELETE /api/alert", handler.DeleteAlert)

	router.HandleFunc("POST /api/job", handler.CreateJob)
	router.HandleFunc("GET /api/jobs", handler.GetJobs)
	// router.HandleFunc("PATCH /api/job", handler.UpdateAlert)
	// router.HandleFunc("DELETE /api/job", handler.DeleteAlert)

	//TODO - API endpoint needed to monitor the mail channel capacity and size

	http.ListenAndServe(":8080", router)
}

func configure() error {

	// Create logger for writing information and error messages.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	appConfig = config.AppWideConfig{
		InfoLog:  infoLog,
		ErrorLog: errLog,
	}

	props := config.ReadConfigFile()
	infoLog.Println("Contents of property file are ", props)

	//create a mail channel
	mailChan := make(chan models.MailData, props.SMTPChannelBufSize)

	//create a local cache
	mailTemplates, err := helpers.CreateTemplateCache()
	if err != nil {
		errLog.Fatal("Cache cannot be created")
	}

	//create mongo connection
	client, err := createMongoConnection(props.MongoURL)
	if err != nil {
		errLog.Fatal("Unable to establish connection to mongo")
	}

	//setup repository
	alertRepo := repository.NewAlertRepository(client.Database(props.MongoDBName), props.AlertCollectionName)
	jobRepo := repository.NewJobRepository(client.Database(props.MongoDBName), props.JobCollectionName)

	appConfig = config.AppWideConfig{
		Properties:        props,
		MailChannel:       mailChan,
		MailTemplateCache: mailTemplates,
		MongoClient:       client,
		AlertRepo:         alertRepo,
		JobRepo:           jobRepo,
		InfoLog:           infoLog,
		ErrorLog:          errLog,
	}

	//pass down config to service layer
	service.SetConfig(&appConfig)

	//start listening to email messages sent to a channel
	service.ListenToMessages()

	//start a cron to send mails on a schedule
	scheduler.StartCRONScheduler(appConfig.Properties.DefaultJobRefresh)
	return nil
}

// createMongoConnection creates a mongo connection
func createMongoConnection(url string) (*mongo.Client, error) {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(url).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		panic(err)
	}
	appConfig.InfoLog.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return client, nil
}
