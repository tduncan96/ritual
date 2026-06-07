package cmd

import (
	"fmt"
	"os"
	"strconv"

	"ritual/internal/db"
	"ritual/internal/exec"
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

var runJob = &cobra.Command{
	Use: "run <id>",
	Short: "run the given job now",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		job, err := db.GetJob(id)
		if err != nil {
			return err
		}
		if err := exec.ExecuteJob(job); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Job ID %d\n successfully started", id)
		return nil
	},
}

var createJob = &cobra.Command{
	Use: "create <job name> <schedule> <host> <commands> <env file>",
	Short: "create job",
	Args: cobra.RangeArgs(4, 5),
	RunE: func(cmd *cobra.Command, args []string) error {
		var newJob db.Job
		newJob.JobName = args[0]
		newJob.Schedule = args[1]
		newJob.Host = args[2]
		newJob.Commands = args[3]
		if len(args) == 5 {
			file, err := os.ReadFile(args[4])
			if err != nil {
				return fmt.Errorf("error opening env file: %w", err)
			}
			newJob.Env = db.EnvStringToMap(string(file))
		}

		id, err := newJob.CreateJob()
		if err != nil {
			return fmt.Errorf("error creating job: %w", err)
		}

		fmt.Printf("job successfully created: ID: %v", id)
		return nil
	},
}

func init() {
	importCmd.AddCommand(importTomlCmd)
	importCmd.AddCommand(importCronCmd)
	rootCmd.AddCommand(importCmd)

	rootCmd.AddCommand(runJob)
	rootCmd.AddCommand(createJob)
}
