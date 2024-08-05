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

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var appConfig *config.AppWideConfig

func SetConfig(c *config.AppWideConfig) {
	appConfig = c
}

func ListenToMessages() {
	go func() {
		for {
			msg := <-appConfig.MailChannel
			SendMailUsingDefault(msg)
		}
	}()
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
		log.Println("Failed to send", err)
	} else {
		log.Println("Sent successfully")
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

func CreateAlert(w http.ResponseWriter, alert *models.Alert) {
	//alert.MigrationDate = time.Now().AddDate(0, 0, 2) //added manually
	//set defaults
	alert.AlertType = "email"
	alert.AlertSchedule = appConfig.Properties.DefaultAlert
	alert.TemplateName = appConfig.Properties.DefaultTemplate
	alert.IndexId = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := appConfig.AlertRepo.Create(ctx, alert)
	if err != nil {
		serverError(w, err)
	}
	//construct maildata model using the alert request and send basic email alert
}

func GetAlerts(w http.ResponseWriter, filter interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list, err := appConfig.AlertRepo.List(ctx, filter)
	if err != nil {
		serverError(w, err)
		return
	}
	data, err := json.Marshal(list)
	if err != nil {
		serverError(w, err)
	}
	if len(data) != 4 {
		w.Write(data)
	} else {
		noDataFound(w)
	}
}

// func GetAlertsByDate(w http.ResponseWriter) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	now := time.Now()
// 	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	appConfig.InfoLog.Println("Current Date is ", currentDate)
// 	list, err := appConfig.AlertRepo.List(ctx, bson.M{
// 		"migrationdate": bson.M{
// 			"$gte": currentDate,
// 			"$lt":  currentDate.AddDate(0, 0, 7),
// 		},
// 	})
// 	if err != nil {
// 		serverError(w, err)
// 		return
// 	}
// 	data, err := json.Marshal(list)
// 	if err != nil {
// 		serverError(w, err)
// 	}
// 	appConfig.InfoLog.Println("value is ", string(data[:]))
// 	if len(data) != 4 {
// 		w.Write(data)
// 	}

// }

func serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	appConfig.ErrorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ClientError(w http.ResponseWriter, status int, err error) {
	appConfig.ErrorLog.Output(2, err.Error())
	http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
}

func noDataFound(w http.ResponseWriter) {
	http.Error(w, "No data found for this range", http.StatusNoContent)
}
