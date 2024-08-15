package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"runtime/debug"
	"strings"
	"time"

	"citi.com/179563_genesis_mail/pkg/config"
	"citi.com/179563_genesis_mail/pkg/models"

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
	appConfig.InfoLog.Printf("Creating %d workers for sending mail", numWorkers)
	// Start worker goroutines to read mailData concurrently and send mails
	for i := 1; i <= numWorkers; i++ {
		//wg.Add(1)
		go func() {
			for {
				msg := <-appConfig.MailChannel
				appConfig.InfoLog.Printf("Worker %d received %v\n", i, msg)
				SendMailUsingDefault(msg)
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

	mailBody, err := fetchMailBody(m)
	if err != nil {
		return
	}

	msg := []byte("To: " + strings.Join(m.To, " ") + "\r\n" +
		"From: " + m.From + "\r\n" +
		"Subject: " + m.Subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		mailBody)

	err = smtp.SendMail(address, auth, m.From, m.To, msg)
	if err != nil {
		appConfig.ErrorLog.Println("Failed to send email", err)
		//TODO write to DB?

	} else {
		appConfig.InfoLog.Println("Sent successfully")
		//TODO Update DB with current time when alert was sent?
	}
}

func fetchMailBody(m models.MailData) (string, error) {
	if m.Template == "" {
		return "", errors.New("noo template name was set in the model")
	} else {
		myCache := appConfig.MailTemplateCache
		t, ok := myCache[m.Template]
		if !ok {
			appConfig.ErrorLog.Println("Not found in cache - Template Looked up value is ", m.Template)
			return "", errors.New("not found in cache")
		}

		buf := new(bytes.Buffer)
		err := t.Execute(buf, m)
		if err != nil {
			appConfig.ErrorLog.Println("Template execution failed ", err)
			return "", err
		}
		return buf.String(), nil
	}
}

func CreateAlert(alert *models.Alert) ([]byte, error) {

	//set defaults
	alert.IndexId = primitive.NewObjectID()
	alert.AlertType = "email"

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	err := appConfig.AlertRepo.Create(ctx, alert)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(alert)
	if err != nil {
		return nil, err
	}

	sendMail(alert, &models.Job{
		TemplateName: appConfig.Properties.DefaultTemplate,
		MailSubject:  "Migration Request Created",
	})

	return data, nil
}

func GetAlerts(filter interface{}) ([]byte, []*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
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

func GetAlertById(id primitive.ObjectID) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	alert, err := appConfig.AlertRepo.GetByID(ctx, id)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, err
	}
	return alert, nil
}

func UpdateAlert(updates *models.Alert) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	err := appConfig.AlertRepo.Update(ctx, updates, updates.IndexId)
	// err := appConfig.AlertRepo.Update(ctx, updates)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, err
	}
	return GetAlertById(updates.IndexId)
}

func DeleteAlert(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	deleteCount, err := appConfig.AlertRepo.Delete(ctx, id)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return err
	}
	if deleteCount == 0 {
		return errors.New("unable to find the Alert Id")
	}
	return nil
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	appConfig.ErrorLog.Output(2, trace)
	w.WriteHeader(http.StatusInternalServerError)
	errResponse := models.ErrorResponse{
		ErrorMessage: err.Error(),
	}
	json.NewEncoder(w).Encode(errResponse)
}

func ClientError(w http.ResponseWriter, status int, err error) {

	appConfig.ErrorLog.Output(3, err.Error())
	w.WriteHeader(http.StatusBadRequest)
	errResponse := models.ErrorResponse{
		ErrorMessage: err.Error(),
	}
	json.NewEncoder(w).Encode(errResponse)
	//http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
}

func NoDataFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	errResponse := models.ErrorResponse{
		ErrorMessage: "No data found",
	}
	json.NewEncoder(w).Encode(errResponse)
	// http.Error(w, "No data found", http.StatusNoContent)
}

func GetJobs(filter interface{}) ([]byte, []*models.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	list, err := appConfig.JobRepo.List(ctx, filter)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, nil, err
	}
	data, err := json.Marshal(list)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, nil, err
	}
	return data, list, nil
}

func CreateJob(job *models.Job) ([]byte, error) {
	job.IndexId = primitive.NewObjectID()
	job.CreatedAt = time.Now()
	//To do - Data validation and check for date range

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	err := appConfig.JobRepo.Create(ctx, job)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, err
	}
	data, err := json.Marshal(job)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, err
	}
	return data, nil
}

func GetJobById(id primitive.ObjectID) (*models.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	job, err := appConfig.JobRepo.GetByID(ctx, id)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, err
	}
	return job, nil
}

func UpdateJob(updates *models.Job) (*models.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	err := appConfig.JobRepo.Update(ctx, updates, updates.IndexId)
	// err := appConfig.JobRepo.Update(ctx, updates)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return nil, err
	}
	return GetJobById(updates.IndexId)
}

func DeleteJob(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.Properties.MongoTimeout)*time.Second)
	defer cancel()
	deleteCount, err := appConfig.JobRepo.Delete(ctx, id)
	if err != nil {
		appConfig.ErrorLog.Println(err)
		return err
	}
	if deleteCount == 0 {
		return errors.New("unable to find the Job Id")
	}
	return nil
}

// StartSecondaryCron starts secondary jobs based on entries from jobs table
// To DO - this has to be made synchronous and thread safe
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
		secondaryCrons.AddFunc(name, func() { setScheduledAlerts(eachJob) })
		appConfig.InfoLog.Println("Now adding cron - ", eachJob.Comments, eachJob)
	}
	secondaryCrons.Start()
	appConfig.InfoLog.Println("Job schedule as seen in DB are ", secondaryCrons.Entries())

	//set application wide config, this may be made available as part of API to update at real time and should be made thread safe
	appConfig.CronJobs = secondaryCrons
	appConfig.JobMap = jobMap
}

func setScheduledAlerts(job *models.Job) {
	appConfig.InfoLog.Printf("Alert: %s at %s\n", job.Comments, time.Now().Format(time.RFC1123))

	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// filter := bson.M{
	// 	"migrationDate": bson.M{
	// 		"$gte": currentDate.AddDate(0, 0, int(job.FromDate)),
	// 		"$lt":  currentDate.AddDate(0, 0, int(job.ToDate)),
	// 	},
	// 	"isReadyToSend": true,
	// }

	filter := bson.M{
		"migrationDate": bson.M{
			"$gte": currentDate.AddDate(0, 0, int(job.StartDate)),
			"$lt":  currentDate.AddDate(0, 0, int(job.EndDate)),
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
	for _, alert := range alerts {
		sendMail(alert, job)
	}
}

// sendMail Takes alert object and constructs a maildata object and sends mail
func sendMail(alert *models.Alert, job *models.Job) {

	content := make(map[string]interface{})
	content["MigrationId"] = alert.MigrationId
	content["Volumes"] = alert.Volumes
	content["MigrationDate"] = alert.MigrationDate.Format(time.RFC822)

	mail := models.MailData{
		To:       alert.DistributionList,
		From:     appConfig.Properties.FromAddress,
		Subject:  job.MailSubject,
		Content:  content,
		Template: job.TemplateName,
	}

	appConfig.MailChannel <- mail
}
