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

func NotifyDateChange(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	migrationId := strings.TrimPrefix(req.URL.Path, "/api/alert/notifyDateChange/")
	var dateChangeRequest models.NotifyMigrationDateChange
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&dateChangeRequest)
	if err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}

	//TODO Update the migration id with isReadyToSend to false
	updatedAlert, err := service.UpdateAlert(&models.Alert{
		IsReadyToSend: dateChangeRequest.IsReadyToSend,
	}, migrationId)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			service.ClientError(w, http.StatusNotFound, errors.New("unable to find the migration id"))
		} else {
			service.ServerError(w, err)
		}
		return
	}

	data, err := json.Marshal(updatedAlert)
	if err != nil {
		service.ServerError(w, err)
		return
	}

	err = service.NotifyDateChange(migrationId, dateChangeRequest)
	if err != nil {
		service.ServerError(w, err)
		return
	}

	w.Write(data)
}

// GetAlert - returns alert matching the migration id
func GetAlert(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	migrationId := strings.TrimPrefix(req.URL.Path, "/api/alert/")
	alert, err := service.GetAlertById(migrationId)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			service.ClientError(w, http.StatusNotFound, errors.New("unable to find the migration id"))
		} else {
			service.ServerError(w, err)
		}
		return
	}
	data, err := json.Marshal(alert)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	w.Write(data)
}

// GetAlerts - returns list of all alerts based on a date range, if no date range specified default date is used. i.e list all alerts whose MD >= current date. Date range should be specified as ?startDate=2024-08-21&endDate=2024-08-24 (YYYY-MM-DD)
func GetAlerts(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	startDateString := req.URL.Query().Get("startDate")
	endDateString := req.URL.Query().Get("endDate")
	now := time.Now()
	var filter primitive.M

	if startDateString == "" || endDateString == "" {
		currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		filter = bson.M{
			"migrationDate": bson.M{
				"$gte": currentDate,
			},
		}
	} else {
		layout := "2006-01-02"
		startDate, err := time.ParseInLocation(layout, startDateString, now.Location())
		if err != nil {
			service.ClientError(w, http.StatusBadRequest, err)
			return
		}
		endDate, err := time.ParseInLocation(layout, endDateString, now.Location())
		if err != nil {
			service.ClientError(w, http.StatusBadRequest, err)
			return
		}
		filter = bson.M{
			"migrationDate": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		}
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

// UpdateAlert - updates an existing alert matching by migrationId
func UpdateAlert(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	migrationId := strings.TrimPrefix(req.URL.Path, "/api/alert/")
	isApproved := req.URL.Query().Get("isApproved")
	isApproved = strings.ToLower(isApproved)
	var updateRequest, oldAlert *models.Alert

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&updateRequest)
	if err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}

	if updateRequest.MigrationId != "" {
		service.ClientError(w, http.StatusBadRequest, errors.New("you cannot update migration id"))
		return
	} else {
		//TODO get alert model before update
		oldAlert, err = service.GetAlertById(migrationId)
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				service.ClientError(w, http.StatusNotFound, errors.New("unable to find the migration id"))
			} else {
				service.ServerError(w, err)
			}
			return
		}
	}

	//Check if a flag value is set in the request, if not use the value already present in Mongo
	// This done to prevent zero value
	if updateRequest.IsReadyToSend == nil {
		updateRequest.IsReadyToSend = oldAlert.IsReadyToSend
	}

	newAlert, err := service.UpdateAlert(updateRequest, migrationId)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			service.ClientError(w, http.StatusNotFound, errors.New("unable to find the migration id"))
		} else {
			service.ServerError(w, err)
		}
		return
	}

	data, err := json.Marshal(newAlert)
	if err != nil {
		service.ServerError(w, err)
		return
	}

	// Check if the request includes a date change, and request parameter has isApproved Flag set to true
	if updateRequest.MigrationDate.String() != "" && isApproved == "true" {
		service.SendApprovedMail(oldAlert, updateRequest.MigrationDate)
	}

	// Check if the request includes only isReadyToSend flag value as true and request parameter has isApproved flag set to false
	if isApproved == "false" && *updateRequest.IsReadyToSend {
		service.SendRejectMail(newAlert)
	}
	//TODO compare the values that have changed in old vs new, and build a dynamic table
	w.Write(data)
}

// DeleteAlert - deletes an alert matching by migrationId
func DeleteAlert(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	migrationId := strings.TrimPrefix(req.URL.Path, "/api/alert/")
	err := service.DeleteAlert(migrationId)
	if err != nil {
		if err.Error() == "unable to find the migration Id" {
			service.ClientError(w, http.StatusNotFound, err)
		} else {
			service.ServerError(w, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateJob - creates a job to be picked up by scheduler for automated emails
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
}

// GetJobs - list all jobs present in the collection
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

// GetJob - returns a job matched by jobId
func GetJob(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobId := strings.TrimPrefix(req.URL.Path, "/api/job/")
	objId, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}
	job, err := service.GetJobById(objId)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			service.ClientError(w, http.StatusBadRequest, errors.New("unable to find the Job Id"))
		} else {
			service.ServerError(w, err)
		}
		return
	}
	data, err := json.Marshal(job)
	if err != nil {
		service.ServerError(w, err)
		return
	}
	w.Write(data)
}

// UpdateJob - Updates a job, matched by jobId
func UpdateJob(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobId := strings.TrimPrefix(req.URL.Path, "/api/job/")
	objId, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		service.ClientError(w, http.StatusBadRequest, err)
		return
	}
	var job models.Job
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&job)
	if err != nil {
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
	w.Write(data)
}

// DeleteJob - Deletes a job, matched by jobId
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
