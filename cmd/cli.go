package cmd

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"slices"

	"ritual/codec"
	"ritual/internal/db"
	"ritual/internal/execute"

	"github.com/spf13/cobra"
)

var dumpPath = os.Getenv("RITUAL_CRON_PATH")
var host string
var crontab bool
var batch bool

var importCmd = &cobra.Command{
	Use:   "import <flags> <file path or directory; if none defualt to $RITUAL_CRON_PATH",
	Short: "Import file contents as new job(s)",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var files []string
		if crontab {
			files = append(files, "crontab")
		} else if len(args) == 1 {
			info, err := os.Stat(args[0])
			switch {
			case err == nil && info.Mode().IsRegular():
				files = append(files, args[0])
			case err == nil && info.IsDir():
				allFiles, err := os.ReadDir(args[0])
				if err != nil {
					return err
				}
				for _, path := range allFiles {
					if path.IsDir() {
						continue
					}
					files = append(files, filepath.Join(args[0], path.Name()))
				}
			default:
				return err
			}
		} else {
			allFiles, err := os.ReadDir(dumpPath)
			if err != nil {
				return err
			}
			for _, path := range allFiles {
				if path.IsDir() {
					continue
				}
				files = append(files, filepath.Join(dumpPath, path.Name()))
			}
		}

		var fileType string
		var content []byte
		for _, file := range files {
			if file == "crontab" {
				fileType = "cron"
				cmd := exec.Command("crontab", "-l")
				out, err := cmd.Output()
				if err != nil {
					return err
				}
				content = out
			} else {
				fileType = strings.Replace(filepath.Ext(file), ".", "", 1)
				out, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				content = out
			}

			defs, err := codec.Codecs[fileType].Unmarshal(content)
			if err != nil {
				return err
			}
			for _, def := range defs {
				job := db.DefToJob(def)
				id, err := job.CreateJob()
				if err != nil {
					fmt.Printf("error creating job: %v", err)
					continue
				}
				fmt.Printf("Job #%d successfully created", id)
			}
		}

		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export <file type> <job id(s); if none export all jobs>",
	Short: "Export job(s) to file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileType := args[0]
		possTypes := slices.Collect(maps.Keys(codec.Codecs))
		if !slices.Contains(possTypes, fileType) {
			err := fmt.Errorf("invalid file type: %v", fileType)
			return err
		}
		
		var jobList []db.Job
		if len(args) == 1 {
			jobs, err := db.GetAllJobs()
			if err != nil {
				return err
			}
			jobList = jobs
		} else {
			var ids []int64
			for _, arg := range args[1:] {
				id, err := strconv.Atoi(arg)
				if err != nil {
					return err
				}
				ids = append(ids, int64(id))
			}
			jobs, err := db.GetJobs(ids)
			if err != nil {
				return err
			}
			jobList = jobs
		}

		var defs []codec.Definition
		for _, job := range jobList {
			defs = append(defs, db.JobToDef(job))
		}

		var content []byte
		if batch {
			blob, err := codec.Codecs[fileType].Marshal(defs)
			if err != nil {
				fmt.Printf("error marshaling job: %v", err)
			}
			content = append(content, blob...)
			if err := os.WriteFile("batch."+fileType, content, 0o644); err != nil {
				fmt.Printf("error writing to file: %v", err)
			}
		} else {
			for _, def := range defs {
				d := []codec.Definition{def}
				blob, err := codec.Codecs[fileType].Marshal(d)
				if err != nil {
					fmt.Printf("error marshaling job: %v", err)
				}
				content = append(content, blob...)
				if !batch {
					if err := os.WriteFile(def.Name+"."+fileType, content, 0o644); err != nil {
						fmt.Printf("error writing to file: %v", err)
					}
					content = []byte{}
				}
			}
		}
		return nil
	},
}

var runJob = &cobra.Command{
	Use:   "run <id>",
	Short: "run a given job now",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		job, err := db.GetJob(int64(id))
		if err != nil {
			return err
		}
		if err := execute.ExecuteJob(job); err != nil {
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
	importCmd.Flags().BoolVarP(&crontab, "crontab", "c", false, "import from crontab")
	importCmd.Flags().StringVarP(&host, "host", "h", "localhost", "import from given host")

	exportCmd.Flags().BoolVarP(&batch, "batch", "b", false, "batch export jobs to file")
	exportCmd.Flags().StringVarP(&host, "host", "h", "localhost", "export to given host")

	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(runJob)
	rootCmd.AddCommand(createJob)
}
