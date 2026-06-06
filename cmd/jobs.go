package cmd

import (
	"fmt"

	"ritual/internal/imports"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import jobs from external source",
}

var importTomlCmd = &cobra.Command{
	Use:   "toml <file>",
	Short: ".toml file job import",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := imports.TomlToJob(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Imported Job ID: %d\n", id)
		return nil
	},
}

var importCronCmd = &cobra.Command{
	Use:   "cron <host>",
	Short: "crontab job imports",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		ids, err := imports.CrontabToJobs(host)
		if err != nil {
			return err
		}
		fmt.Printf("job ids created: %v", ids)
		return nil
	},
}

func init() {
	importCmd.AddCommand(importTomlCmd)
	importCmd.AddCommand(importCronCmd)
	rootCmd.AddCommand(importCmd)
}
