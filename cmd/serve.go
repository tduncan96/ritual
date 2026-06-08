package cmd

import (
	"fmt"

	"ritual/internal/db"
	"ritual/internal/exec"
	"ritual/internal/web"

	robfig "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Ritual cron runner and web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cron := robfig.New()
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
			} else {
				fmt.Printf("job %v added to cron as entry %v", job.JobId, entryId)
			}
		}

		cron.Start()
		web.Start()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
