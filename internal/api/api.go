package api

import (
	"encoding/json"
	"net/http"
	"ritual/internal/ops"
)

func createJobHandler(w http.ResponseWriter, r *http.Request) {
	var request ops.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := request.CreateJobCall()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&response)
}

func publishEventHandler(w http.ResponseWriter, r *http.Request) {
	var request ops.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := request.PublishEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&response)
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/jobs/new", createJobHandler)
	mux.HandleFunc("POST /api/publish", publishEventHandler)
}
