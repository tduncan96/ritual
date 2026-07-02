package cron

import (
	"encoding/json"

	"ritual/bus"
	"ritual/internal/db"
	"ritual/internal/logger"
	"ritual/internal/run"

	robfig "github.com/robfig/cron/v3"
)

var log = logger.For("cron")

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
					log.Error().
						Err(err).
						Job(logger.Execute, job).
						Msg("error executing job")
				}
			})
			if err != nil {
				log.Error().
					Err(err).
					Job(logger.Serve, job).
					Msg("could not add job to cron runner")
			} else {
				cr.Lookup[job.JobId] = entryId
				log.Info().
					Job(logger.Serve, job).
					Msgf("job added to cron as entry #%v", entryId)
			}
		}
	}
}

func (cr *CronRunner) UpdateRunner(ids []int64) error {
	jobs, err := db.GetJobs(ids)
	if err != nil {
		return err
	}

	for _, id := range ids {
		cr.Cron.Remove(cr.Lookup[id])
		delete(cr.Lookup, id)
	}
	cr.AddJobs(jobs)

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
				log.Error().
					Err(err).
					Msg("error unmarshaling event payload")
				continue
			}
			switch event.Method {
			case bus.POST:
				if err := cr.UpdateRunner(ids); err != nil {
					log.Error().
						Err(err).
						Msg("error updating cron runner from event payload")
				} else {
					log.Info().
						Msg("cron runner jobs updated")
				}
			case bus.DELETE:
				cr.RemoveRunnerJob(ids)
			}
		}
	}
}
