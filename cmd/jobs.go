package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"ritual/bus"
	"ritual/codec"
	"ritual/internal/db"
	"ritual/internal/ops"
	"ritual/internal/run"
	"ritual/internal/srv"

	"github.com/spf13/cobra"
)

var dumpPath = os.Getenv("RITUAL_CRON_PATH")

var host string
var crontab bool
var batch bool

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "execute commands against jobs",
}

var importJob = &cobra.Command{
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
		var jobList []db.Job
		for _, file := range files {
			if file == "crontab" {
				fileType = "cron"

				runner := run.Runner{
					Job: db.Job{
						Commands: "crontab -l",
						Host:     &host,
					},
				}
				if err := runner.ResolveTarget(); err != nil {
					return err
				}
				out, code, err := runner.RunCommand()
				slog.Info(fmt.Sprintf("crontab exited with code %v", code))
				if err != nil {
					slog.Error(fmt.Sprintf("crontab exited with code %v", code), "error", err)
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

			possTypes := slices.Collect(maps.Keys(codec.Codecs))
			if !slices.Contains(possTypes, fileType) {
				err := fmt.Errorf("invalid file type: %v", fileType)
				return err
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
				jobList = append(jobList, job)
				fmt.Fprintf(cmd.OutOrStdout(), "Job #%d successfully created\n", id)
			}
		}

		if err := publishToDaemon(jobList, bus.POST); err != nil {
			return err
		}

		return nil
	},
}

var exportJob = &cobra.Command{
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

		if !batch || fileType == "cron" {
			for _, def := range defs {
				d := []codec.Definition{def}
				blob, err := codec.Codecs[fileType].Marshal(d)
				if err != nil {
					fmt.Printf("error marshaling job %v: %v", def.Name, err)
					continue
				}
				if err := os.WriteFile(def.Name+"."+fileType, blob, 0o600); err != nil {
					fmt.Printf("error writing to file for job %v: %v", def.Name, err)
					continue
				}
			}
		} else {
			blob, err := codec.Codecs[fileType].Marshal(defs)
			if err != nil {
				return fmt.Errorf("error marshaling job: %v", err)
			}
			if err := os.WriteFile("batch."+fileType, blob, 0o600); err != nil {
				return fmt.Errorf("error writing to file: %v", err)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Jobs successfully exported!\n")

		if err := publishToDaemon(jobList, bus.GET); err != nil {
			return err
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
		runner := run.Runner{Job: job}
		if err := runner.ExecuteJob(); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Job #%d successfully started\n", id)

		if err := publishToDaemon([]db.Job{job}, bus.GET); err != nil {
			return err
		}

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
			Host:     &args[2],
			Commands: args[3],
			Status:   true,
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
		fmt.Fprintf(cmd.OutOrStdout(), "job successfully created: ID: %d\n", id)

		newJob.JobId = id
		if err := publishToDaemon([]db.Job{newJob}, bus.POST); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	importJob.Flags().BoolVarP(&crontab, "crontab", "c", false, "import from crontab")
	importJob.Flags().StringVarP(&host, "host", "H", "localhost", "use host other than localhost")

	exportJob.Flags().BoolVarP(&batch, "batch", "b", false, "batch export jobs to file")

	jobCmd.AddCommand(importJob)
	jobCmd.AddCommand(exportJob)
	jobCmd.AddCommand(runJob)
	jobCmd.AddCommand(createJob)

	rootCmd.AddCommand(jobCmd)
}

func publishToDaemon(jobs []db.Job, method bus.Method) error {
	var ids []int64
	for _, job := range jobs {
		ids = append(ids, job.JobId)
	}
	payload, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	requestBody, err := json.Marshal(ops.RequestBody{
		Events: []bus.Event{{
			SubList: bus.Database,
			Method:  method,
			Payload: payload,
		}}})
	if err != nil {
		return err
	}
	client := srv.NewSocketClient()
	response, err := client.Post("http://unix/api/publish", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		slog.Warn("could not notify daemon", "error", err)
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("bad server response %s: %s", response.Status, body)
	}

	return nil
}
