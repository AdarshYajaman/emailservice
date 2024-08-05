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
	service.CreateAlert(w, &alertRequest)
	fmt.Fprintf(w, "Executing POST on Alerts for %s %s %s", alertRequest.MigrationId, alertRequest.Volumes, alertRequest.AlertType)
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
	service.GetAlerts(w, filter)
	fmt.Fprintf(w, "Executing Get on Alerts")
}

func UpdateAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Updates on Alerts")
}

func DeleteAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Delete on Alerts")
}
