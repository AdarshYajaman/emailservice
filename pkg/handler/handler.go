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
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}
	alertRequest.AlertType = "email"
	alertRequest.MigrationDate = time.Now().AddDate(0, 0, 2)
	service.CreateAlert(w, &alertRequest)
	fmt.Fprintf(w, "Executing POST on Alerts for %s %s %s", alertRequest.MigrationId, alertRequest.Volumes, alertRequest.AlertType)
}

func GetAlert(w http.ResponseWriter, req *http.Request) {
	service.GetAlertsByDate(w)
	fmt.Fprintf(w, "Executing Get on Alerts")
}

func UpdateAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Updates on Alerts")
}

func DeleteAlert(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Executing Delete on Alerts")
}
