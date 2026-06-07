package cron

import (
	"fmt"
	"ritual/internal/db"
	"ritual/internal/exec"

	robfig "github.com/robfig/cron/v3"
)

func PopulateCron(cron *robfig.Cron) error {
	allJobs, err := db.GetAllJobs()
	if err != nil {
		return err
	}
	for _, job := range allJobs {
		entryId, err := cron.AddFunc(job.Schedule, func() {
			if err := exec.ExecuteJob(job); err != nil {
				fmt.Printf("error executing job #%v - %v: %v", job.JobId, job.JobName, err)
			}
		})
		if err != nil {
			fmt.Printf("could not add job %v to cron: %v", job.JobId, err)
		}
		fmt.Printf("job %v added to cron as entry %v", job.JobId, entryId)
	}
	return nil
}
