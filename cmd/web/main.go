package main

import (
	"103-EmailService/pkg/config"
	"103-EmailService/pkg/handler"
	"103-EmailService/pkg/helpers"
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/service"
	"fmt"
	"log"
	"net/http"

	"github.com/robfig/cron/v3"
)

var AppConfig config.AppWideConfig

func main() {

	//Read .properties file and configure the map
	props, err := helpers.ReadPropertiesFile("./pkg/config/application.properties")
	if err != nil {
		log.Fatal("Unable to locate and parse the property file, failed with error - ", err)
	}
	log.Println("Contents of property file are ", props)

	//Create an application wide config to be used which can be passed down to any packages
	mailChan := make(chan models.MailData)
	mailTemplates, err := helpers.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cache cannot be created")
	}

	AppConfig = config.AppWideConfig{
		Properties:    props,
		MailChannel:   mailChan,
		TemplateCache: mailTemplates,
	}
	service.SetConfig(&AppConfig)

	// service.SendMailDefault()

	//Below implementation uses https://github.com/xhit/go-simple-mail/tree/master
	defer close(AppConfig.MailChannel)

	service.ListenToMessages()

	var contentMap = map[string]string{
		"name": "Adarsh",
	}

	mailData := models.MailData{
		To:       []string{"to@test.com"},
		From:     "from@test.com",
		Subject:  "Test email",
		Content:  contentMap,
		Template: "simple.page.tmpl",
	}

	AppConfig.MailChannel <- mailData
	// createData(&appConfig)

	router := http.NewServeMux()
	router.HandleFunc("GET /api/alerts", handler.GetAlerts)
	router.HandleFunc("POST /api/alerts", handler.CreateAlert)
	router.HandleFunc("PATCH /api/alerts", handler.UpdateAlerts)
	router.HandleFunc("DELETE /api/alerts", handler.DeleteAlerts)

	http.ListenAndServe(":8080", router)

	c := cron.New()
	c.AddFunc("* * * * *", func() { fmt.Println("Every 1 min on the half hour") })
	c.Start()
}
