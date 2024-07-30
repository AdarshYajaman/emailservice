package handler

import (
	"103-EmailService/pkg/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func CreateAlert(w http.ResponseWriter, req *http.Request) {
	var alertRequest models.CreateAlertRequest
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&alertRequest)
	if err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Executing POST on Alerts for %s %s %s", alertRequest.MigrationId, alertRequest.Volumes, alertRequest.AlertType)
}

func GetAlerts(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Get on Alerts")
}

func UpdateAlerts(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Updates on Alerts")
}

func DeleteAlerts(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Delete on Alerts")
}
