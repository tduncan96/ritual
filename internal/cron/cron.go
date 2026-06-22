package cron

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"ritual/bus"
	"ritual/internal/db"
	"ritual/internal/run"

	robfig "github.com/robfig/cron/v3"
)

type CronRunner struct {
	Cron   *robfig.Cron
	Lookup map[int64]robfig.EntryID
}

func MakeRunner() (*CronRunner, error) {
	cr := CronRunner{
		Cron:   robfig.New(),
		Lookup: make(map[int64]robfig.EntryID),
	}

	allJobs, err := db.GetAllJobs()
	if err != nil {
		return nil, err
	}

	cr.AddJobs(allJobs)

	return &cr, nil
}

func (cr *CronRunner) AddJobs(jobs []db.Job) {
	for _, job := range jobs {
		if job.Status {
			entryId, err := cr.Cron.AddFunc(job.Schedule, func() {
				runner := run.Runner{Job: job}
				if err := runner.ExecuteJob(); err != nil {
					slog.Error(fmt.Sprintf("error executing job #%v - %v: %v", job.JobId, job.JobName, err), "error", err)
				}
			})
			if err != nil {
				slog.Error(fmt.Sprintf("could not add job #%v to cron runner", job.JobId), "error", err, "job", job.JobId)
			} else {
				slog.Info(fmt.Sprintf("job #%v added to cron as entry %v", job.JobId, entryId), "job", job.JobId)
			}
			cr.Lookup[job.JobId] = entryId
		}
	}
}

func (cr *CronRunner) UpdateRunner(ids []int64) error {
	jobs, err := db.GetJobs(ids)
	if err != nil {
		return err
	}

	cr.Cron.Stop()
	for _, id := range ids {
		cr.Cron.Remove(cr.Lookup[id])
		delete(cr.Lookup, id)
	}
	cr.AddJobs(jobs)
	cr.Cron.Start()

	slog.Info("cron runner jobs updated", "ids", ids)
	return nil
}

func (cr *CronRunner) RemoveRunnerJob(ids []int64) {
	for _, id := range ids {
		cr.Cron.Remove(cr.Lookup[id])
		delete(cr.Lookup, id)
	}
}

func CronSubscription(cr *CronRunner, subLists ...bus.SubList) {
	ch := bus.GlobalBus.Subscribe(subLists...)
	defer bus.GlobalBus.Unsubscribe(ch, subLists...)
	for event := range ch {
		switch event.SubList {
		case bus.LifeCycle:
			switch event.Method {
			case bus.PUT:
				cr.Cron.Start()
			case bus.DELETE:
				cr.Cron.Stop()
			}
		case bus.Database:
			var ids []int64
			if err := json.Unmarshal(event.Payload, &ids); err != nil {
				slog.Error("error unmarshaling event payload", "error", err)
				return
			}
			switch event.Method {
			case bus.POST:
				if err := cr.UpdateRunner(ids); err != nil {
					slog.Error("error updating cron runner from event payload", "error", err, "ids", ids)
				} else {
					slog.Info("cron runner jobs updated", "ids", ids)
				}
			case bus.DELETE:
				cr.RemoveRunnerJob(ids)
			}
		}
	}
}
