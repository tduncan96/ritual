package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	robfig "github.com/robfig/cron/v3"
)

// Job
type Job struct {
	JobId     int    `db:"JobId"`
	JobName   string `db:"JobName" toml:"name"`
	Schedule  string `db:"Schedule" toml:"schedule"`
	Host      string `db:"Host" toml:"host"`
	Commands  string `db:"Commands" toml:"commands"`
	Env       EnvMap `db:"Env"`
	JobType   string `db:"JobType" toml:"type"`
	JobStatus string `db:"JobStatus"`
	Created   string `db:"Created"`
	Updated   string `db:"Updated"`
	LastRun   string `db:"LastRun"`
	NextRun   string `db:"NextRun"`
}

func (j *Job) CreateJob() (int64, error) {
	result, err := DB.NamedExec(
		`INSERT INTO Jobs (JobName, Schedule, Host, Commands, Env) 
		VALUES (:JobName, :Schedule, :Host, :Commands, :Env)`,
		j,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	j.JobId = int(id)
	return id, nil
}

func (j *Job) UpdateJob() error {
	if _, err := DB.NamedExec(
		`UPDATE Jobs SET
				JobName  = :JobName,
				Schedule = :Schedule,
				Host     = :Host,
				Commands = :Commands,
				Env      = :Env,
				JobType  = :JobType,
				JobStatus = :JobStatus,
				Updated  = datetime('now'),
				LastRun = :LastRun,
				NextRun = :NextRun
				WHERE JobId = :JobId`,
		j,
	); err != nil {
		return err
	}
	fmt.Println("job #%d successfully update", j.JobId)
	return nil
}

func (j *Job) CalcNextRun() error {
	next, err := robfig.ParseStandard(j.Schedule)
	if err != nil {
		return err
	}
	j.NextRun = next.Next(time.Now()).Format(SqlTimeFormat)
	if err := j.UpdateJob(); err != nil {
		return err
	}
	return nil
}

func DeleteJob(id int) (int64, error) {
	result, err := DB.Exec("DELETE FROM jobs WHERE JobId = ?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetJob(id int) (Job, error) {
	var job Job
	err := DB.Get(&job, "SELECT * FROM Jobs WHERE JobId = ?", id)
	if err != nil {
		return Job{}, err
	}
	return job, nil
}

func GetJobs(ids []int) ([]Job, error) {
	var jobs []Job
	for _, id := range ids {
		job, err := GetJob(id)
		if err != nil {
			fmt.Printf("error getting job JobId: %v", id)
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func GetAllJobs() ([]Job, error) {
	var jobs []Job
	var ids []int
	err := DB.Select(&ids, "SELECT JobId FROM Jobs")
	if err != nil {
		return []Job{}, err
	}
	jobs, err = GetJobs(ids)
	if err != nil {
		return []Job{}, err
	}
	return jobs, nil
}

// EnvMap
type EnvMap map[string]string

var _ driver.Valuer = EnvMap{}
var _ sql.Scanner = (*EnvMap)(nil)

func EnvMapToString(envMap map[string]string) (envString string) {
	if len(envMap) > 0 {
		var envStrings []string
		for key, value := range envMap {
			envLine := key + "=" + value
			envStrings = append(envStrings, envLine)
		}
		envString = strings.Join(envStrings, "\n")
	} else {
		envString = ""
	}
	return envString
}
func (em EnvMap) Value() (driver.Value, error) {
	return EnvMapToString(em), nil
}

func EnvStringToMap(envString string) (envMap map[string]string) {
	envMap = make(map[string]string)
	for _, line := range strings.Split(envString, "\n") {
		if line == "" {
			continue
		}
		envExp := strings.SplitN(line, "=", 2)
		if len(envExp) != 2 {
			continue
		}
		envMap[envExp[0]] = envExp[1]
	}
	return envMap
}
func (em *EnvMap) Scan(src any) error {
	*em = EnvStringToMap(src.(string))
	return nil
}
