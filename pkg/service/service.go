package service

import (
	"103-EmailService/pkg/config"
	"103-EmailService/pkg/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var appConfig *config.AppWideConfig

func SetConfig(c *config.AppWideConfig) {
	appConfig = c
}

func ListenToMessagesOld() {
	go func() {
		for {
			msg := <-appConfig.MailChannel
			SendMailUsingDefault(msg)
		}
	}()
}

// ListenToMessages creates workers (routines) that can read and send mails concurrently
func ListenToMessages() {
	numWorkers := appConfig.Properties.SMTPWorkers
	//var wg sync.WaitGroup

	// Start worker goroutines to read mailData concurrently and send mails
	for i := 1; i <= numWorkers; i++ {
		//wg.Add(1)

		go func() {
			for {
				// msg := <-appConfig.MailChannel
				appConfig.InfoLog.Printf("Worker %d received %v\n", i, "test")
				// SendMailUsingDefault(msg)
				//wg.Done()
			}
		}()

		//wg.Wait()
	}
}

func SendMailUsingDefault(m models.MailData) {
	props := appConfig.Properties
	address := props.SMTPHost + ":" + props.SMTPPort
	auth := smtp.PlainAuth("", "", "", props.SMTPHost)

	msg := []byte("To: " + strings.Join(m.To, " ") + "\r\n" +
		"From: " + m.From + "\r\n" +
		"Subject: " + m.Subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		fetchMailBody(m))

	err := smtp.SendMail(address, auth, m.From, m.To, msg)
	if err != nil {
		appConfig.ErrorLog.Println("Failed to send", err)
	} else {
		appConfig.InfoLog.Println("Sent successfully")
	}
}

func fetchMailBody(m models.MailData) string {
	if m.Template == "" {
		return ""
	} else {
		myCache := appConfig.MailTemplateCache
		t, ok := myCache[m.Template]
		if !ok {
			log.Println("Not found in cache")
			return ""
		}

		buf := new(bytes.Buffer)
		err := t.Execute(buf, m)
		if err != nil {
			log.Println("Template execution failed ", err)
			return ""
		}
		return buf.String()
	}
}

func CreateAlert(alert *models.Alert) ([]byte, error) {

	//set defaults
	alert.AlertType = "email"
	alert.IndexId = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := appConfig.AlertRepo.Create(ctx, alert)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(alert)
	if err != nil {
		return nil, err
	}

	// construct maildata model using the alert request and send basic email alert, this has to be made async
	content := make(map[string]interface{})
	content["MigrationId"] = alert.MigrationId
	content["Volumes"] = alert.Volumes
	content["MigrationDate"] = alert.MigrationDate.Format(time.RFC822)

	mail := models.MailData{
		To:       []string{"test@test.com"},
		From:     appConfig.Properties.FromAddress,
		Subject:  "Migration Request Created",
		Content:  content,
		Template: appConfig.Properties.DefaultTemplate,
	}

	appConfig.MailChannel <- mail

	return data, nil
}

func GetAlerts(filter interface{}) ([]byte, []*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list, err := appConfig.AlertRepo.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(list)
	if err != nil {
		return nil, nil, err
	}
	return data, list, nil
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	appConfig.ErrorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ClientError(w http.ResponseWriter, status int, err error) {
	appConfig.ErrorLog.Output(2, err.Error())
	http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
}

func NoDataFound(w http.ResponseWriter) {
	http.Error(w, "No data found for this range", http.StatusNoContent)
}

func GetJobs(filter interface{}) ([]byte, []*models.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list, err := appConfig.JobRepo.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(list)
	if err != nil {
		return nil, nil, err
	}
	return data, list, nil
}

func CreateJob(job *models.Job) ([]byte, error) {
	job.IndexId = primitive.NewObjectID()
	job.CreatedAt = time.Now()
	//To do - Data validation and check for date range

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := appConfig.JobRepo.Create(ctx, job)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// StartSecondaryCron starts secondary jobs based on entries from jobs table
func StartSecondaryCron() {
	var secondaryCrons = appConfig.CronJobs

	// Check secondary cron if entries exists remove them, in order to start with latest schedule as defined in jobs table
	if secondaryCrons == nil {
		// appConfig.InfoLog.Println("This will be nil during startup, create a new cron")
		secondaryCrons = cron.New()
	} else {
		secondaryCrons.Stop()
		for _, eachEntry := range secondaryCrons.Entries() {
			appConfig.InfoLog.Println("Full Entries before ", secondaryCrons.Entries())
			secondaryCrons.Remove(cron.EntryID(eachEntry.ID))
			appConfig.InfoLog.Println("Full Entries After ", secondaryCrons.Entries())
		}
	}

	//Update schedule map with new entries from DB, this is created to be passed to service layer
	_, jobList, err := GetJobs(bson.M{})
	if err != nil {
		appConfig.ErrorLog.Printf("Unable to search for jobs - check mongo collection name : %v Error : %v \n", appConfig.Properties.JobCollectionName, err)
		return
	}
	jobMap := make(map[string]*models.Job)
	for _, eachJob := range jobList {
		name := eachJob.CronExpression
		jobMap[name] = eachJob
		secondaryCrons.AddFunc(name, func() { sendScheduledAlert(eachJob) })
		appConfig.InfoLog.Println("Now adding cron - ", eachJob.Comments)
	}
	secondaryCrons.Start()
	appConfig.InfoLog.Println("Full Entries reflecting DB is ", secondaryCrons.Entries())

	//set application wide config, this may be made available as part of API to update at real time and should be made thread safe
	appConfig.CronJobs = secondaryCrons
	appConfig.JobMap = jobMap

}

func sendScheduledAlert(job *models.Job) {
	appConfig.InfoLog.Printf("Alert: %s at %s\n", job.Comments, time.Now().Format(time.RFC1123))

	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// filter := bson.M{
	// 	"migrationdate": bson.M{
	// 		"$gte": currentDate.AddDate(0, 0, int(job.FromDate)),
	// 		"$lt":  currentDate.AddDate(0, 0, int(job.ToDate)),
	// 	},
	// 	"isreadytosend": true,
	// }

	filter := bson.M{
		"migrationdate": bson.M{
			"$gte": currentDate.AddDate(0, 0, int(job.FromDate)),
			"$lt":  currentDate.AddDate(0, 0, int(job.ToDate)),
		},
	}

	_, alerts, err := GetAlerts(filter)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return
	}

	sendMails(alerts, job)

}

func sendMails(alerts []*models.Alert, job *models.Job) {

}
