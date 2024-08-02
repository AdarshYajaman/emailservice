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

	"go.mongodb.org/mongo-driver/bson"
)

var appConfig *config.AppWideConfig

func SetConfig(c *config.AppWideConfig) {
	appConfig = c
}

func ListenToMessages() {
	go func() {
		for {
			msg := <-appConfig.MailChannel
			// SendMailUsingGoMail(msg)
			SendMailUsingDefault(msg)
		}
	}()
}

func SendMailUsingDefault(m models.MailData) {
	smtpDetails := appConfig.Properties
	user := smtpDetails["smtp.username"]
	password := smtpDetails["smtp.password"]
	host := smtpDetails["smtp.host"]
	port := smtpDetails["smtp.port"]
	address := host + ":" + port
	auth := smtp.PlainAuth("", user, password, host)

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
		myCache := appConfig.TemplateCache
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

func CreateAlert(w http.ResponseWriter, car *models.Alert) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := appConfig.AlertRepo.Create(ctx, car)
	if err != nil {
		serverError(w, err)
	}
}

func GetAlerts(w http.ResponseWriter) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list, err := appConfig.AlertRepo.List(ctx, bson.M{})
	if err != nil {
		serverError(w, err)
		return
	}
	data, err := json.Marshal(list)
	if err != nil {
		serverError(w, err)
	}
	w.Write(data)
}

func GetAlertsByDate(w http.ResponseWriter) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	currentTime := time.Now()
	list, err := appConfig.AlertRepo.List(ctx, bson.M{
		"migrationdate": bson.M{
			"$gte": currentTime,
			"$lt":  currentTime.AddDate(0, 0, 7),
		},
	})
	if err != nil {
		serverError(w, err)
		return
	}
	data, err := json.Marshal(list)
	if err != nil {
		serverError(w, err)
	}
	appConfig.InfoLog.Println("value is ", string(data[:]))
	if data != nil {
		w.Write(data)
	}

}

func serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	appConfig.ErrorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ClientError(w http.ResponseWriter, status int, err error) {
	appConfig.InfoLog.Output(2, err.Error())
	http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
}
