package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jakecoffman/gorunner/executor"
	"github.com/jakecoffman/gorunner/models"
	"io/ioutil"
	"net/http"
	"strconv"
)

type addJobPayload struct {
	Name string `json:"name"`
}

type addTaskToJobPayload struct {
	Task string `json:"task"`
}

func ListJobs(w http.ResponseWriter, r *http.Request) {
	jobList := models.GetJobList()

	w.Write([]byte(models.Json(jobList)))
}

func AddJob(w http.ResponseWriter, r *http.Request) {
	jobList := models.GetJobList()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload addJobPayload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.Name == "" {
		http.Error(w, "Please provide a 'name'", http.StatusBadRequest)
		return
	}

	err = models.Append(jobList, models.Job{Name: payload.Name, Status: "New"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func GetJob(w http.ResponseWriter, r *http.Request) {
	jobList := models.GetJobList()

	vars := mux.Vars(r)
	job, err := models.Get(jobList, vars["job"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(bytes)
}

func DeleteJob(w http.ResponseWriter, r *http.Request) {
	jobList := models.GetJobList()

	vars := mux.Vars(r)
	job, err := models.Get(jobList, vars["job"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = models.Delete(jobList, job.ID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AddTaskToJob(w http.ResponseWriter, r *http.Request) {
	jobList := models.GetJobList()

	vars := mux.Vars(r)
	job, err := models.Get(jobList, vars["job"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	j := job.(models.Job)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload addTaskToJobPayload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.Task == "" {
		http.Error(w, "Please provide a 'task' to add to "+j.Name, http.StatusBadRequest)
		return
	}
	j.AppendTask(payload.Task)
	models.Update(jobList, j)

	w.WriteHeader(201)
}

func RemoveTaskFromJob(w http.ResponseWriter, r *http.Request) {
	jobList := models.GetJobList()

	vars := mux.Vars(r)
	job, err := models.Get(jobList, vars["job"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	j := job.(models.Job)

	taskPosition, err := strconv.Atoi(vars["task"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	j.DeleteTask(taskPosition)
	models.Update(jobList, j)
}

func JobTrigger(w http.ResponseWriter, r *http.Request) {
	jobList := models.GetJobList()

	vars := mux.Vars(r)
	job, err := models.Get(jobList, vars["job"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	j := job.(models.Job)

	if r.Method == "DELETE" {
		j.DeleteTrigger(vars["trigger"])
		models.Update(jobList, j)
	} else if r.Method == "POST" {
		trigger := r.FormValue("trigger")
		j.AppendTrigger(trigger)
		triggerList := models.GetTriggerList()
		t, _ := models.Get(triggerList, trigger)
		executor.AddTrigger <- t.(models.Trigger)
		models.Update(jobList, j)
	}

}
