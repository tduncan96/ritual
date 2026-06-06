package cmd

import (
	"os"

	"ritual/internal/cron"
	"ritual/internal/web"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve <service>",
	Short: "Serve rituals services",
}

var serveWebCmd = &cobra.Command{
	Use:   "web",
	Short: "Start Ritual web server",
	Run: func(cmd *cobra.Command, args []string) {
		port := os.Getenv("RITUAL_PORT")
		if port == "" {
			port = "8080"
		}

		s := &web.Server{DB: Database}

		s.Start(port)
	},
}

var serveCronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Start Ritual cron service",
	Run: func(cmd *cobra.Command, args []string) {
		cron.MasterCron.Start() // this is wrong
	},
}

func init() {
	serveCmd.AddCommand(serveWebCmd)
	serveCmd.AddCommand(serveCronCmd)
	rootCmd.AddCommand(serveCmd)
}
