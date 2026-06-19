package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"ritual/internal/bus"
	"ritual/internal/cron"
	"ritual/internal/srv"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Ritual cron runner and web server",
	RunE: func(cmd *cobra.Command, args []string) error {

		cron, err := cron.MakeRunner()
		if err != nil {
			return err
		}

		bus.MakeBus()
		srv.MakeMux()

		go bus.CronSubscription(cron, bus.LifeCycle, bus.Database)

		cron.Cron.Start()
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
