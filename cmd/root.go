package cmd

import (
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

var Database *sqlx.DB

var rootCmd = &cobra.Command{
	Use:   "ritual",
	Short: "Automation scheduler",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
