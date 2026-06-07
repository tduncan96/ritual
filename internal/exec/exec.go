package exec

import (
	"errors"
	"os"
	"os/exec"
	"time"

	"ritual/internal/db"
)

func ExecuteJob(job db.Job) error {
	var errs []error
	run := db.Run{
		JobId:   &job.JobId,
		JobName: job.JobName,
		Host:    job.Host,
	}

	start := time.Now()
	cmd := exec.Command("sh", "-c", job.Commands)
	cmd.Env = os.Environ()
	for k, v := range job.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			run.ExitCode = ee.ExitCode()
		} else {
			run.ExitCode = -1
			errs = append(errs, err)
		}
	} else {
		run.ExitCode = 0
	}

	end := time.Now()

	run.StartTime = db.TimeStamp(start)
	run.EndTime = db.TimeStamp(end)
	run.Duration = int64(end.Sub(start))
	run.Logs = string(out)

	_, err = run.CreateRun()
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// This will need to be changed to accomodate ssh later on
