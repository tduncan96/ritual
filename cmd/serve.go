package cmd

import (
	"os"

	"ritual/internal/web"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Ritual web server",
	Run: func(cmd *cobra.Command, args []string) {
		Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func Serve() {
	port := os.Getenv("RITUAL_PORT")
	if port == "" {
		port = "8080"
	}

	s := &web.Server{DB: Database}

	s.Start(port)
}
