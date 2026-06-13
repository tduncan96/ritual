package cmd

import (
	"fmt"

	"ritual/internal/api"
	"ritual/internal/db"
	"ritual/internal/execute"
	"ritual/internal/web"

	robfig "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Ritual cron runner and web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		bus := api.NewBus()
		api.Bus = bus

		allJobs, err := db.GetAllJobs()
		if err != nil {
			return err
		}

		cron := robfig.New()
		var Lookup = make(map[int64]robfig.EntryID)
		for _, job := range allJobs {
			if job.Status {
				entryId, err := cron.AddFunc(job.Schedule, func() {
					var runner execute.Runner
					if job.Host == "localhost" {
						runner = execute.LocalRunner{}
					} else {
						runner = execute.RemoteRunner{}
					}
					if err := runner.ExecuteJob(job); err != nil {
						fmt.Printf("error executing job #%v - %v: %v", job.JobId, job.JobName, err)
					}
				})
				if err != nil {
					fmt.Printf("could not add job %v to cron: %v", job.JobId, err)
				} else {
					fmt.Printf("job %v added to cron as entry %v", job.JobId, entryId)
				}
				Lookup[job.JobId] = entryId
			}
		}

		go api.Subscription(api.Logging, api.Shutdown, api.DBWrites, api.Cron)

		cron.Start()
		web.Start()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
