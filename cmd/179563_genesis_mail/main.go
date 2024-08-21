package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"citi.com/179563_genesis_mail/pkg/config"
	"citi.com/179563_genesis_mail/pkg/handler"
	"citi.com/179563_genesis_mail/pkg/helpers"
	"citi.com/179563_genesis_mail/pkg/models"
	"citi.com/179563_genesis_mail/pkg/repository"
	"citi.com/179563_genesis_mail/pkg/scheduler"
	"citi.com/179563_genesis_mail/pkg/service"

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
	router.HandleFunc("POST /api/alert/notifyDateChange/{migrationId}", handler.NotifyDateChange)
	router.HandleFunc("GET /api/alert/{migrationId}", handler.GetAlert)
	router.HandleFunc("PATCH /api/alert/{migrationId}", handler.UpdateAlert)
	router.HandleFunc("DELETE /api/alert/{migrationId}", handler.DeleteAlert)

	router.HandleFunc("GET /api/jobs", handler.GetJobs)
	router.HandleFunc("POST /api/job", handler.CreateJob)
	router.HandleFunc("GET /api/job/{jobId}", handler.GetJob)
	router.HandleFunc("PATCH /api/job/{jobId}", handler.UpdateJob)
	router.HandleFunc("DELETE /api/job/{jobId}", handler.DeleteJob)

	//TODO - API endpoint to monitor mail channel capacity and size
	//TODO - API endpoint to update job

	srv := &http.Server{
		ReadTimeout:  0 * time.Second,
		WriteTimeout: time.Duration(appConfig.Properties.APITimeOut) * time.Second,
		Addr:         ":" + appConfig.Properties.APIPort,
		Handler:      router,
	}
	log.Println(srv.ListenAndServe())
}

func configure() error {
	// Create logger for writing information and error messages.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	appConfig = config.AppWideConfig{
		InfoLog:  infoLog,
		ErrorLog: errLog,
	}

	props := helpers.ReadConfigFile()
	infoLog.Println("Contents of property file are ", props)

	//create a mail channel
	mailChan := make(chan models.MailData, props.SMTPChannelBufSize)

	//create a local cache
	mailTemplates, err := helpers.CreateTemplateCache()
	if err != nil {
		errLog.Fatal("Cache cannot be created")
		return err
	}

	//create mongo connection
	client, err := createMongoConnection(props.MongoURL)
	if err != nil {
		errLog.Fatal("Unable to establish connection to mongo")
		return err
	}

	//setup repository
	alertRepo := repository.NewAlertRepository(client.Database(props.MongoDBName), props.AlertCollectionName)
	jobRepo := repository.NewJobRepository(client.Database(props.MongoDBName), props.JobCollectionName)
	// alertRepo := repository.NewCustomRepository[*models.Alert](client.Database(props.MongoDBName), props.AlertCollectionName)
	// jobRepo := repository.NewCustomRepository[*models.Job](client.Database(props.MongoDBName), props.JobCollectionName)

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
