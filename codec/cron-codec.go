package codec

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"os/exec"
	"regexp"
	"strings"

	robfig "github.com/robfig/cron/v3"
)

type CronCodec struct{}

var envRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*\s*=`)

func (c CronCodec) Marshal(defs []Definition) (blob []byte, err error) {
	var buf bytes.Buffer
	for _, def := range defs {
		fmt.Fprintf(&buf, "## name: %v\n", def.Name)
		for key, value := range def.Env {
			fmt.Fprintf(&buf, "%s=%s\n", key, value)
		}
		stat := ""
		if def.Status == false {
			stat = "## "
		}
		fmt.Fprintf(&buf, "%s%s %s\n\n", stat, def.Schedule, def.Commands)
	}
	return buf.Bytes(), nil
}

func (c CronCodec) Unmarshal(blob []byte) (defs []Definition, err error) {
	localHost := exec.Command("hostname")
	out, err := localHost.Output()
	if err != nil {
		return []Definition{}, fmt.Errorf("error getting stdout of 'hostname': %w", err)
	}
	host := strings.TrimSpace(string(out))

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(blob))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []Definition{}, err
	}

	var name string
	var status bool
	env := make(map[string]string)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		var sched, cmd string
		switch {

		// ex. @every 5m /usr/.local/bin/script.sh -l
		case strings.HasPrefix(fields[0], "@every"):
			sched, cmd = strings.Join(fields[:2], " "), strings.Join(fields[2:], " ")

		// ex. @monthly /usr/.local/bin/script.sh -l
		case strings.HasPrefix(fields[0], "@"):
			sched, cmd = fields[0], strings.Join(fields[1:], " ")

		// ex. ## name: Example Job Name
		case fields[1] == "name:":
			name = strings.Join(fields[2:], "")
			continue

		// ex. 0 2 * * * /usr/.local/bin/script.sh -l
		default:
			if len(fields) > 6 && fields[0] == "##" {
				sched, cmd = strings.Join(fields[0:6], " "), strings.Join(fields[6:], " ")
				status = false
			} else if len(fields) > 5 {
				sched, cmd = strings.Join(fields[:5], " "), strings.Join(fields[5:], " ")
			} else {
				continue
			}
		}

		if _, err := robfig.ParseStandard(sched); err != nil {
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

		lineEnv := make(map[string]string, len(env))
		maps.Copy(lineEnv, env)
		hash := GetHash(host, sched, cmd, lineEnv)
		if name == "" {
			name = strings.Join([]string{host, "crontab", hash}, "_")
		}
		def := Definition{
			Name:     name,
			Host:     host,
			Schedule: sched,
			Commands: cmd,
			Env:      lineEnv,
			Hash:     hash,
			Status:   status,
		}
		defs = append(defs, def)

		name = ""
		status = true
	}
	return defs, nil
}
