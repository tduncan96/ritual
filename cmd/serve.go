package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ritual/internal/bus"
	"ritual/internal/db"
	"ritual/internal/run"
	"ritual/internal/srv"

	robfig "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Ritual cron runner and web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		bus.MakeBus()
		srv.MakeMux()

		allJobs, err := db.GetAllJobs()
		if err != nil {
			return err
		}

		cron := robfig.New()
		var Lookup = make(map[int64]robfig.EntryID)
		for _, job := range allJobs {
			if job.Status {
				entryId, err := cron.AddFunc(job.Schedule, func() {
					var runner run.Runner
					if job.Host == "localhost" {
						runner = run.LocalRunner{}
					} else {
						runner = run.RemoteRunner{}
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

		go bus.Subscription(bus.Shutdown, bus.Database)

		cron.Start()
		go srv.SocketServe()
		go srv.WebServe()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		<-ctx.Done()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
