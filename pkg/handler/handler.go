package handler

import (
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/service"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func CreateAlert(w http.ResponseWriter, req *http.Request) {
	var alertRequest models.Alert
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&alertRequest)
	if err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}
	data, err := service.CreateAlert(&alertRequest)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		service.ServerError(w, err)
	}
	// fmt.Fprintf(w, "Executing POST on Alerts for %s %s %s", alertRequest.MigrationId, alertRequest.Volumes, alertRequest.AlertType)
}

func GetAlert(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	filter := bson.M{
		"migrationdate": bson.M{
			"$gte": currentDate,
			"$lt":  currentDate.AddDate(0, 0, 7),
		},
		"isreadytosend": true,
	}
	data, err := service.GetAlerts(filter)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	if len(data) != 4 {
		w.Write(data)
	} else {
		service.NoDataFound(w)
	}
}

func UpdateAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Updates on Alerts")
}

func DeleteAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Delete on Alerts")
}

func CreateSchedule(w http.ResponseWriter, req *http.Request) {
	var jobRequest models.Job
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&jobRequest)
	if err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}
	data, err := service.CreateJob(&jobRequest)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		service.ServerError(w, err)
	}
	// fmt.Fprintf(w, "Executing POST on Alerts for %s %s %s", jobRequest.MigrationId, jobRequest.Volumes, jobRequest.AlertType)
}

func GetSchedule(w http.ResponseWriter, req *http.Request) {

	filter := bson.M{}
	data, _, err := service.GetJobs(filter)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	if len(data) != 4 {
		w.Write(data)
	} else {
		service.NoDataFound(w)
	}
}
