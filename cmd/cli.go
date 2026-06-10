package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ritual/codec"
	"ritual/internal/db"
	"ritual/internal/exec"

	"github.com/spf13/cobra"
)

var dumpPath = os.Getenv("RITUAL_CRON_PATH")
var host string
var crontab bool

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import file contents as new job(s)",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var files []string
		var errs []error

		if len(args) == 1 {
			files = []string{args[0]}
		} else {
			allFiles, err := os.ReadDir(dumpPath)
			if err != nil {
				return err
			}
			for _, path := range allFiles {
				if path.IsDir() {
					continue
				}
				files = append(files, path.Name())
			}
		}

		for _, file := range files {
			fileType := strings.Replace(filepath.Ext(file), ".", "", 1)

			content, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			defs, err := codec.Codecs["toml"].Unmarshal(content)
			if err != nil {
				return err
			}
			var errs []error
			for _, def := range defs {
				job := db.DefToJob(def)
				id, err := job.CreateJob()
				if err != nil {
					errs = append(errs, fmt.Errorf("error creating job: %w", err))
				}
				fmt.Printf("Job #%d successfully created", id)
			}
		}

		return errors.Join(errs...)
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export jobs from internal source",
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
	importCmd.Flags().StringVarP(&host, "host", "h", "localhost", "import from given host")
	exportCmd.Flags().StringVarP(&host, "host", "h", "localhost", "export to given host")

	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(runJob)
	rootCmd.AddCommand(createJob)
}
