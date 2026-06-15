package ops

import (
	"errors"
	"fmt"
	"ritual/codec"
	"ritual/internal/bus"
	"ritual/internal/db"
)

type RequestBody struct {
	JobIds []int64            `json:"job_ids,omitempty"`
	Defs   []codec.Definition `json:"definitions,omitempty"`
	Events []bus.Event        `json:"events,omitempty"`
	Host   string             `json:"host,omitempty"`
}

type ResponseBody struct {
	JobIds []int64 `json:"job_ids"`
	Error  error   `json:"error"`
}

func CreateJobCall(request RequestBody) (response ResponseBody, err error) {
	if request.Defs == nil {
		return response, errors.New("no job definitions passed in request")
	} else {
		var errs []error
		for _, def := range request.Defs {
			job := db.DefToJob(def)
			id, err := job.CreateJob()
			if err != nil {
				cErr := fmt.Errorf("error creating job '%v': %w", def.Name, err)
				errs = append(errs, cErr)
			}
			response.JobIds = append(response.JobIds, id)
		}
		response.Error = errors.Join(errs...)
	}

	var logEntry []byte
	if response.JobIds != nil {
		log := fmt.Sprintf("job id(s) successfully created: %v.", response.JobIds)
		logEntry = append(logEntry, log...)
	} else {
		log := fmt.Sprint("no new jobs created")
		logEntry = append(logEntry, log...)
	}
	if response.Error != nil {
		log := []byte(fmt.Sprintf("\nerrors encountered: %v", response.Error))
		logEntry = append(logEntry, log...)
	}

	logging := bus.Event{
		SubList: bus.Logging,
		Method:  bus.PUT,
		Payload: logEntry,
	}
	dbWrite := bus.Event{
		SubList: bus.DBWrites,
		Method:  bus.POST,
		Payload: []byte{},
	}
	bus.GlobalBus.Publish(dbWrite, logging)

	return response, nil
}

func PublishEvents(request RequestBody) (response ResponseBody, err error) {
	if request.Events == nil {
		return response, errors.New("no events passed in request")
	} else {
		bus.GlobalBus.Publish(request.Events...)
	}
	return response, nil
}
