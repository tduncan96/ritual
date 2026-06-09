package codec

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"hash/fnv"

	robfig "github.com/robfig/cron/v3"
)

type CronCodec struct{}

var envRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*\s*=`)

func jobName(host, schedule, command string) string {
    h := fnv.New32a()
    h.Write([]byte(schedule))
    h.Write([]byte("\x00"))
    h.Write([]byte(command))
    return fmt.Sprintf("%s_%08x", host, h.Sum32())
}

func (c CronCodec) Marshal(def Definition) (blob []byte, err error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "## name: %v\n", def.Name)
	for key, value := range def.Env {
		fmt.Fprintf(&buf, "%s=%s\n", key, value)
	}
	stat := ""
	if def.Status == false {
		stat = "## "
	}
	fmt.Fprintf(&buf, "%s%s %s\n", stat, def.Schedule, def.Commands)

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

	env := make(map[string]string)
	var name string
	var status bool
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
		case strings.HasPrefix(fields[0], "## "):
			if fields[1] == "name:" {
				name = strings.Join(fields[2:], "")
			}
		default:
			if len(fields) < 6 {
				// Move on to next check
			} else {
				sched, cmd = strings.Join(fields[:5], " "), strings.Join(fields[5:], " ")
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
		if name == "" {
			name = strings.Join([]string{host, "crontab", jobName(host, sched, cmd)}, "_")
		}
		def := Definition{
			Name:     name,
			Host:     host,
			Schedule: sched,
			Commands: cmd,
			Env:      lineEnv,
			Status:   true,
		}
		defs = append(defs, def)
		name = ""
	}
	return defs, nil
}
