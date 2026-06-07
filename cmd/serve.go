package cmd

import (
	"ritual/internal/cron"
	"ritual/internal/web"

	robfig "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Ritual web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cronRunner := robfig.New()
		if err := cron.PopulateCron(cronRunner); err != nil {
			return err
		}
		cronRunner.Start()
		web.Serve()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
