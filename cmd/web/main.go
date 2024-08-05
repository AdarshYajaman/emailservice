package main

import (
	"103-EmailService/pkg/config"
	"103-EmailService/pkg/handler"
	"103-EmailService/pkg/helpers"
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/repository"
	"103-EmailService/pkg/service"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
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

	router.HandleFunc("POST /api/schedule", handler.CreateSchedule)
	router.HandleFunc("GET /api/schedule", handler.GetSchedule)

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
	alertRepo := repository.NewAlertRepository(client.Database(props.MongoDBName), props.AlertCollectionName)
	jobRepo := repository.NewJobRepository(client.Database(props.MongoDBName), props.JobCollectionName)

	// start cron job
	// cronJobs := startCron()
	// infoLog.Println("Entries are", cronJobs.Entries())

	// for eachEntry := range cronJobs.Entries() {
	// 	infoLog.Println(cronJobs.Entry(cron.EntryID(eachEntry)))
	// }

	appConfig = config.AppWideConfig{
		Properties:        props,
		MailChannel:       mailChan,
		MailTemplateCache: mailTemplates,
		MongoClient:       client,
		AlertRepo:         alertRepo,
		JobRepo:           jobRepo,
		InfoLog:           infoLog,
		ErrorLog:          errLog,
		// CronJobs:          cronJobs,
	}

	service.SetConfig(&appConfig)

	test()

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

type scheduler struct {
}

func (s scheduler) Run() {
	var secondaryCrons = appConfig.CronJobs

	// Check secondary cron if entries exists remove them, in order to start with latest schedule
	if secondaryCrons == nil {
		appConfig.InfoLog.Println("This will be nil during startup, create a new cron")
		secondaryCrons = cron.New()
	} else {
		secondaryCrons.Stop()
		for _, eachEntry := range secondaryCrons.Entries() {
			appConfig.InfoLog.Println("Full Entries before ", secondaryCrons.Entries())
			secondaryCrons.Remove(cron.EntryID(eachEntry.ID))
			appConfig.InfoLog.Println("Full Entries After ", secondaryCrons.Entries())
		}
		// secondaryCrons.Stop()
	}

	//Update schedule map with new entries from DB
	_, jobList, err := service.GetJobs(bson.M{})
	if err != nil {
		appConfig.ErrorLog.Println("Unable to search for jobs - check mongo collection name ", appConfig.Properties.JobCollectionName, err)
	}
	jobMap := make(map[string]*models.Job)
	for _, eachJob := range jobList {
		name := eachJob.CronExpression
		jobMap[name] = eachJob
		secondaryCrons.AddFunc(name, func() { sendAlert2(eachJob) })
		appConfig.InfoLog.Println("Now adding cron - ", eachJob.Comments)
	}
	secondaryCrons.Start()
	appConfig.InfoLog.Println("Full Entries reflecting DB is ", secondaryCrons.Entries())

	//set application wide config
	appConfig.CronJobs = secondaryCrons
	appConfig.JobMap = jobMap
}

// This function runs nighly to look for any updates in the jobs table and creates a job schedule accordingly
func test() {

	//Start a nighly cron
	primaryCron := cron.New()
	s := scheduler{}
	primaryCron.AddJob("*/5 * * * *", s)
	primaryCron.Start()

	s.Run()
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
	appConfig.InfoLog.Printf("Alert: %s at %s\n", message, time.Now().Format(time.RFC1123))
}

func sendAlert2(job *models.Job) {
	appConfig.InfoLog.Printf("Alert: %s at %s\n", job.Comments, time.Now().Format(time.RFC1123))
}
