package service

import (
	"103-EmailService/pkg/config"
	"103-EmailService/pkg/models"
	"bytes"
	"context"
	"log"
	"net/smtp"
	"strings"
	"time"
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

func CreateAlert(car *models.Alert) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	log.Println("request is ", car)
	err := appConfig.AlertRepo.Create(ctx, car)
	if err != nil {
		log.Println("Unable to insert this document ", err)
	}
}
