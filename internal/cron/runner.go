package cron

import (
	"errors"
	"fmt"
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
		home, _ := os.UserHomeDir()
		hostKeyCallback, err := knownhosts.New(filepath.Join(home, ".ssh", "known_hosts"))

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
	var envLine strings.Builder
	for k, v := range r.Job.Env {
		value := "'" + strings.ReplaceAll(v, "'", `'\''`) + "'"
		fmt.Fprintf(&envLine, "export %s=%s", k, value)
	}
	envLine.WriteString(r.Job.Commands)

	var errs []error
	if r.Client == nil {
		cmd := exec.Command("sh", "-c", r.Job.Commands)
		out, err = cmd.CombinedOutput()
	} else {
		session, err := r.Client.NewSession()
		if err != nil {
			code = -1
			errs = append(errs, err)
		}
		defer session.Close()
		defer r.Client.Close()
		out, err = session.CombinedOutput(r.Job.Commands)
	}
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			code = ee.ExitCode()
		} else {
			code = -1
			errs = append(errs, err)
		}
	} else {
		code = 0
	}
	return out, code, errors.Join(errs...)
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
	return err
}