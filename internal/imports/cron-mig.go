package imports

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/robfig/cron/v3"

	"ritual/internal/db"
)

var envRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*\s*=`)

func CrontabToJobs(host string) (ids []int64, err error) {
	localHost := exec.Command("hostname")
	out, err := localHost.Output()
	if err != nil {
		return []int64{}, fmt.Errorf("error getting stdout of 'hostname': %w", err)
	}
	if host != strings.TrimSpace(string(out)) { //
		exec.Command("ssh", host) // I'll need to go back and actually do something with this
	}

	crontab := exec.Command("crontab", "-l")
	out, err = crontab.Output()
	if err != nil {
		return []int64{}, err
	}
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []int64{}, err
	}

	env := make(map[string]string)
	i := 1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		var sched, cmd string
		switch {		
		// ex. @every 5m /usr/.local/bin/script.sh
		case strings.HasPrefix(fields[0], "@every"):
			sched, cmd = strings.Join(fields[:2], " "), strings.Join(fields[2:], " ")
		// ex. @monthly /usr/.local/bin/script.sh
		case strings.HasPrefix(fields[0], "@"):
			sched, cmd = fields[0], strings.Join(fields[1:], " ")
		// ex. 0 * * * * /usr/.local/bin/script.sh
		default:
			if len(fields) < 6 {
				continue
			}
			sched, cmd = strings.Join(fields[:5], " "), strings.Join(fields[5:], " ")
		}

		if _, err := cron.ParseStandard(sched); err != nil {
			if envRe.MatchString(line) == true {
				envExp := strings.SplitN(line, "=", 2)
				key := envExp[0]
				value := envExp[1]
				env[key] = value
				continue
			} else {
				fmt.Printf("error parsing line: %v", line)
				continue
			}
		}

		lineEnv := make(db.EnvMap, len(env))
		for k, v := range env {
			lineEnv[k] = v
		}
		num := strconv.Itoa(i)
		job := db.Job{
			JobName:  strings.Join([]string{host, "crontab", num}, "_"),
			Host:     host,
			JobType:  "Bash",
			Schedule: sched,
			Commands: cmd,
			Env:      lineEnv,
		}
		id, err := job.CreateJob()
		if err != nil {
			return []int64{}, err
		}
		ids = append(ids, id)
		fmt.Printf("job %v created", strconv.Itoa(int(id)))
		i++
	}
	return ids, nil
}
