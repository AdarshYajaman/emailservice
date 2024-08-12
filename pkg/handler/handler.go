package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"citi.com/179563_genesis_mail/pkg/models"
	"citi.com/179563_genesis_mail/pkg/service"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateAlert - creates an alert in mongo and sends email notification
func CreateAlert(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
		return
	}
}

func GetAlert(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	alertId := strings.TrimPrefix(req.URL.Path, "/api/alert/")
	objId, err := primitive.ObjectIDFromHex(alertId)
	if err != nil {
		service.ClientError(w, http.StatusNotFound, err)
		return
	}
	alert, err := service.GetAlertById(objId)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	data, err := json.Marshal(alert)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	w.Write(data)
}

// GetAlerts - gets list of all alerts based on a date range
// TODO add date filter in the URL, if not specified use default
func GetAlerts(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	filter := bson.M{
		"migrationDate": bson.M{
			"$gte": currentDate,
			"$lt":  currentDate.AddDate(0, 0, 7),
		},
		"isReadyToSend": true,
	}
	data, _, err := service.GetAlerts(filter)
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

// UpdateAlert - updates an existing alert
func UpdateAlert(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	alertId := strings.TrimPrefix(req.URL.Path, "/api/alert/")
	objId, err := primitive.ObjectIDFromHex(alertId)
	if err != nil {
		service.ClientError(w, http.StatusNotFound, err)
		return
	}
	var alert models.Alert
	if err := json.NewDecoder(req.Body).Decode(&alert); err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}
	alert.IndexId = objId
	value, err := service.UpdateAlert(&alert)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			service.ClientError(w, http.StatusNotFound, errors.New("unable to find the Alert Id"))
		} else {
			service.ServerError(w, err)
		}
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		service.ServerError(w, err)
		return
	}

	//length of data will be 4 when data value is null - may be a duplicate
	if len(data) == 4 {
		service.NoDataFound(w)
		return
	}
	w.Write(data)
}

func DeleteAlert(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	alertId := strings.TrimPrefix(req.URL.Path, "/api/alert/")
	objId, err := primitive.ObjectIDFromHex(alertId)
	if err != nil {
		service.ClientError(w, http.StatusNotFound, err)
		return
	}
	err = service.DeleteAlert(objId)

	if err != nil {
		if err.Error() == "unable to find the Alert Id" {
			service.ClientError(w, http.StatusNotFound, err)
		} else {
			service.ServerError(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CreateJob(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
		return
	}
	// fmt.Fprintf(w, "Executing POST on Alerts for %s %s %s", jobRequest.MigrationId, jobRequest.Volumes, jobRequest.AlertType)
}

func GetJobs(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

func GetJob(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobId := strings.TrimPrefix(req.URL.Path, "/api/job/")
	objId, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		service.ClientError(w, http.StatusNotFound, err)
		return
	}
	job, err := service.GetJobById(objId)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	data, err := json.Marshal(job)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	w.Write(data)
}

func UpdateJob(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobId := strings.TrimPrefix(req.URL.Path, "/api/job/")
	objId, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		service.ClientError(w, http.StatusNotFound, err)
		return
	}
	var job models.Job
	if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}
	job.IndexId = objId
	value, err := service.UpdateJob(&job)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			service.ClientError(w, http.StatusNotFound, errors.New("unable to find the Job Id"))
		} else {
			service.ServerError(w, err)
		}
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	//length of data will be 4 when date value is null
	if len(data) == 4 {
		service.NoDataFound(w)
		return
	}

	// TODO if job id is invalid, throw not found error

	w.Write(data)
}

func DeleteJob(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobId := strings.TrimPrefix(req.URL.Path, "/api/job/")
	objId, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		service.ClientError(w, http.StatusNotFound, err)
		return
	}
	err = service.DeleteJob(objId)
	if err != nil {
		if err.Error() == "unable to find the Job Id" {
			service.ClientError(w, http.StatusNotFound, err)
		} else {
			service.ServerError(w, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
