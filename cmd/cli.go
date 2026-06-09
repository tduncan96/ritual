package cmd

import (
	"fmt"
	"os"
	"strconv"

	"ritual/internal/db"
	"ritual/internal/exec"

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
		if err := jobio.TomlToJob(args[0]); err != nil {
			return err
		}
		return nil
	},
}

var exportTomlCmd = &cobra.Command{
	Use: "export <job id>",
	Short: "Export job to .toml file",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		if err := jobio.JobToToml(id); err != nil {
			return err
		}
		return nil
	},
}

var importCronCmd = &cobra.Command{
	Use:   "cron <host>",
	Short: "crontab job jobio",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		ids, err := jobio.CrontabToJobs(host)
		if err != nil {
			return err
		}
		fmt.Printf("job ids created: %v", ids)
		return nil
	},
}

var runJob = &cobra.Command{
	Use:   "run <id>",
	Short: "run the given job now",
	Args:  cobra.ExactArgs(1),
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
	Use:   "create <job name> <schedule> <host> <commands> <env file>",
	Short: "create job",
	Args:  cobra.RangeArgs(4, 5),
	RunE: func(cmd *cobra.Command, args []string) error {
		newJob := db.Job{
			JobName:  args[0],
			Schedule: args[1],
			Host:     args[2],
			Commands: args[3],
		}
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
