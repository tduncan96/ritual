package run

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"ritual/internal/db"
)

type Runner struct {
	Job    db.Job
	Client *ssh.Client
}

func (r Runner) ExecuteJob() error {
	start := time.Now()
	run := db.Run{
		JobId:     &r.Job.JobId,
		JobName:   r.Job.JobName,
		Host:      r.Job.Host,
		StartTime: db.TimeStamp(start),
	}

	if err := r.resolveTarget(); err != nil {
		return err
	}

	out, code, err := r.runCommand()
	run.Logs = string(out)
	run.ExitCode = int64(code)

	end := time.Now()
	run.EndTime = db.TimeStamp(end)
	run.Duration = int64(end.Sub(start))

	id, err := run.CreateRun()
	if err != nil {
		slog.Error("error creating run record", "error", err)
	} else {
		slog.Info(fmt.Sprintf("job #%v successfully ran and recorded.", r.Job.JobId), "job_id", r.Job.JobId, "run_id", id)
	}
	return err
}

func (r *Runner) resolveTarget() error {
	switch r.Job.Host {
	case "":
		return fmt.Errorf("invalid host")
	case "localhost":
		return nil
	default:
		host, err := db.GetHost(r.Job.Host)
		if err != nil {
			return err
		}
		keyBytes, err := os.ReadFile(host.KeyPath)
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return err
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		hostKeyCallback, err := knownhosts.New(filepath.Join(home, ".ssh", "known_hosts"))
		if err != nil {
			return err
		}

		cfg := &ssh.ClientConfig{
			User:            host.User,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: hostKeyCallback,
			Timeout:         10 * time.Second,
		}

		addr := net.JoinHostPort(host.Address, strconv.Itoa(int(host.Port)))
		client, err := ssh.Dial("tcp", addr, cfg)
		if err != nil {
			return err
		}
		r.Client = client
		return nil
	}
}

func (r *Runner) runCommand() (out []byte, code int, err error) {
	var cmdLine strings.Builder
	for k, v := range r.Job.Env {
		value := "'" + strings.ReplaceAll(v, "'", `'\''`) + "'"
		fmt.Fprintf(&cmdLine, "export %s=%s\n", k, value)
	}
	cmdLine.WriteString(r.Job.Commands)

	var errs []error
	if r.Client == nil {
		cmd := exec.Command("sh", "-c", cmdLine.String())
		out, err = cmd.CombinedOutput()
		if err != nil {
			if ee, ok := errors.AsType[*exec.ExitError](err); ok {
				code = ee.ExitCode()
			} else {
				code = -1
				errs = append(errs, err)
			}
		}
	} else {
		session, sErr := r.Client.NewSession()
		if sErr != nil {
			return nil, -1, sErr
		}
		defer session.Close()
		defer r.Client.Close()
		out, err = session.CombinedOutput(cmdLine.String())
		if err != nil {
			if ee, ok := errors.AsType[*ssh.ExitError](err); ok {
				code = ee.ExitStatus()
			} else {
				code = -1
				errs = append(errs, err)
			}
		}
	}

	return out, code, errors.Join(errs...)
}
