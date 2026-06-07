package exec

import (
	"os"
	"os/exec"
	"strings"

	"ritual/internal/db"
)

func ExecuteJob(job db.Job) error {
	commands := strings.Split(job.Commands, " ")
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Env = os.Environ()
	for k, v := range job.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
