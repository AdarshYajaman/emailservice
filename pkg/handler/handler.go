package handler

import (
	"103-EmailService/pkg/models"
	"103-EmailService/pkg/service"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func CreateAlert(w http.ResponseWriter, req *http.Request) {
	var alertRequest models.Alert
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&alertRequest)
	if err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	alertRequest.AlertType = "email"
	alertRequest.MigrationDate = time.Now().Add(1*time.Hour + 48)
	service.CreateAlert(&alertRequest)
	fmt.Fprintf(w, "Executing POST on Alerts for %s %s %s", alertRequest.MigrationId, alertRequest.Volumes, alertRequest.AlertType)
}

func GetAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Get on Alerts")
}

func UpdateAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Updates on Alerts")
}

func DeleteAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Delete on Alerts")
}
