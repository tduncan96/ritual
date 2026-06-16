package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"ritual/internal/bus"
	"ritual/internal/db"
)

type RequestBody struct {
	Jobs   []db.Job    `json:"jobs,omitempty"`
	Events []bus.Event `json:"events,omitempty"`
	Host   string      `json:"host,omitempty"`
}

type Result struct {
	JobId   int64  `json:"job_id,omitempty"`
	JobName string `json:"job_name"`
	Code    int    `json:"code"`
	Error   string `json:"error,omitempty"`
}

type ResponseBody struct {
	Results []Result `json:"results"`
}

func (request *RequestBody) CreateJobCall() (response ResponseBody, err error) {
	if len(request.Jobs) == 0 {
		err = errors.New("no job definitions passed in request")
		slog.Error("no job definitions passed in request")
		return response, err
	}

	for _, job := range request.Jobs {
		var result Result
		result.JobName = job.JobName
		id, err := job.CreateJob()
		if err != nil {
			result.Code = 1
			result.Error = fmt.Sprintf("error creating job '%v': %v", result.JobName, err)
			slog.Warn("error creating job", "job_name", result.JobName, "error", err)
		} else {
			result.JobId = id
			result.Code = 0
			slog.Info("job successfully created", "job_id", result.JobId, "job_name", result.JobName)
		}
		response.Results = append(response.Results, result)
	}

	var newJobIds []int64
	for _, result := range response.Results {
		if result.Code == 0 {
			newJobIds = append(newJobIds, result.JobId)
		}
	}

	if len(newJobIds) > 0 {
		payload, err := json.Marshal(newJobIds)
		if err != nil {
			return response, err
		}
		dbWrite := bus.Event{
			SubList: bus.DBWrites,
			Method:  bus.POST,
			Payload: payload,
		}
		bus.GlobalBus.Publish(dbWrite)
	}
	
	return response, nil
}

func (request *RequestBody) PublishEvents() (response ResponseBody, err error) {
	if request.Events == nil {
		return response, errors.New("no events passed in request")
	} else {
		bus.GlobalBus.Publish(request.Events...)
	}
	return response, nil
}
