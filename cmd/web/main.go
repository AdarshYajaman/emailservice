package main

import (
	"103-EmailService/pkg/config"
	"103-EmailService/pkg/handler"
	"103-EmailService/pkg/helpers"
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/repository"
	"103-EmailService/pkg/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
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
	router.HandleFunc("GET /api/alert", handler.GetAlert)
	router.HandleFunc("POST /api/alert", handler.CreateAlert)
	router.HandleFunc("PATCH /api/alert", handler.UpdateAlert)
	router.HandleFunc("DELETE /api/alert", handler.DeleteAlert)
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

	mailChan := make(chan models.MailData)

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
	alertRepo := repository.NewAlertRepository(client.Database(props.MongoDBName), props.MongoCollectionName)

	// start cron job
	cronJobs := startCron()
	infoLog.Println("Entries are", cronJobs.Entries())

	for eachEntry := range cronJobs.Entries() {
		infoLog.Println(cronJobs.Entry(cron.EntryID(eachEntry)))
	}

	appConfig = config.AppWideConfig{
		Properties:        props,
		MailChannel:       mailChan,
		MailTemplateCache: mailTemplates,
		MongoClient:       client,
		AlertRepo:         alertRepo,
		InfoLog:           infoLog,
		ErrorLog:          errLog,
		// CronJobs:          cronJobs,
	}

	service.SetConfig(&appConfig)

	return nil
}

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

func startCron() *cron.Cron {
	c := cron.New()

	// Define a list of cron expressions and corresponding messages
	jobs := map[string]string{
		"*/1 * * * *": "Alert every minute",
		"*/5 * * * *": "Alert every 5 minutes",
		"0 0 * * 0":   "Alert every Sunday at midnight",
	}

	// Add jobs to the cron scheduler
	for spec, message := range jobs {
		msg := message // create a new variable to avoid closure capture issue
		_, err := c.AddFunc(spec, func() { sendAlert(msg) })

		if err != nil {
			log.Fatalf("Error adding job: %v", err)
		}
	}

	// Start the cron scheduler
	c.Start()

	return c
}

func sendAlert(message string) {
	fmt.Printf("Alert: %s at %s\n", message, time.Now().Format(time.RFC1123))
}
