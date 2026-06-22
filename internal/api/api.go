package api

import (
	"encoding/json"
	"log/slog"
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
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		slog.Error("error writing http body", "error", err)
	}}

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
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		slog.Error("error writing http body", "error", err)
	}
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/jobs/new", createJobHandler)
	mux.HandleFunc("POST /api/publish", publishEventHandler)
}
