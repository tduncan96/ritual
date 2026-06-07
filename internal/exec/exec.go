package exec

import (
	"os/exec"
	"strings"

	"ritual/internal/db"
)

func ExecuteJob(job db.Job) error {
	commands := strings.Split(job.Commands, " ")
	envVars := strings.Split(db.EnvMapToString(job.Env), "\n")

	cmd := exec.Command(commands[0], commands[1:]...)

	if envVars != nil {
		cmd.Env = envVars
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
