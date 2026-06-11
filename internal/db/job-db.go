package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"ritual/codec"

	robfig "github.com/robfig/cron/v3"
)

// Job
type Job struct {
	JobId    int64  `db:"JobId"`
	JobName  string `db:"JobName"`
	Schedule string `db:"Schedule"`
	Host     string `db:"Host"`
	Commands string `db:"Commands"`
	Env      envMap `db:"Env"`
	Hash     string `db:"Hash"`
	Status   bool   `db:"Status"`
	Created  string `db:"Created"`
	Updated  string `db:"Updated"`
	LastRun  string `db:"LastRun"`
	NextRun  string `db:"NextRun"`
}

func (j *Job) CreateJob() (int64, error) {
	j.Hash = codec.GetHash(j.Host, j.Schedule, j.Commands, j.Env)

	result, err := DB.NamedExec(
		`INSERT INTO Jobs (JobName, Schedule, Host, Commands, Env, Hash) 
		VALUES (:JobName, :Schedule, :Host, :Commands, :Env, :Hash)
		ON CONFLICT (Hash) DO NOTHING`,
		j,
	)
	if err != nil {
		return 0, err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if num == 0 {
		var collision Job
		var errs []error
		getErr := DB.Get(&collision, `SELECT JobId, JobName FROM Jobs WHERE Hash = ?`, j.Hash)
		if getErr != nil {
			errs = append(errs, getErr)
		}
		qErr := fmt.Errorf("collision with Job #%v - %v", collision.JobId, collision.JobName)
		errs = append(errs, qErr)
		return 0, errors.Join(errs...)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	j.JobId = id
	return id, nil
}

func (j *Job) UpdateJob() error {
	j.Hash = codec.GetHash(j.Host, j.Schedule, j.Commands, j.Env)
	if _, err := DB.NamedExec(
		`UPDATE Jobs SET
				JobName  = :JobName,
				Schedule = :Schedule,
				Host     = :Host,
				Commands = :Commands,
				Env      = :Env,
				Hash     = :Hash,
				Status   = :Status,
				Updated  = datetime('now'),
				LastRun  = :LastRun,
				NextRun  = :NextRun
				WHERE JobId = :JobId`,
		j,
	); err != nil {
		return err
	}

	fmt.Printf("job #%d successfully updated\n", j.JobId)
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

func DeleteJob(id int64) (int64, error) {
	result, err := DB.Exec("DELETE FROM Jobs WHERE JobId = ?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetJob(id int64) (Job, error) {
	var job Job
	err := DB.Get(&job, "SELECT * FROM Jobs WHERE JobId = ?", id)
	if err != nil {
		return Job{}, err
	}
	return job, nil
}

func GetJobs(ids []int64) ([]Job, error) {
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
	err := DB.Select(&jobs, "SELECT * FROM Jobs")
	if err != nil {
		return []Job{}, err
	}
	return jobs, nil
}

// envMap
type envMap map[string]string

var _ driver.Valuer = envMap{}
var _ sql.Scanner = (*envMap)(nil)

func EnvMapToString(envMap map[string]string) (envString string) {
	if len(envMap) > 0 {
		var envStrings []string
		for _, key := range slices.Sorted(maps.Keys(envMap)) {
			envLine := strings.Join([]string{key, envMap[key]}, "=")
			envStrings = append(envStrings, envLine)
		}
		envString = strings.Join(envStrings, "\n")
	}
	return envString
}
func (em envMap) Value() (driver.Value, error) {
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
func (em *envMap) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("TimeStamp.Scan: expected string, got %T", src)
	}
	*em = EnvStringToMap(s)
	return nil
}

func DefToJob(def codec.Definition) Job {
	job := Job{
		JobName:  def.Name,
		Schedule: def.Schedule,
		Host:     def.Host,
		Commands: def.Commands,
		Env:      def.Env,
		Status:   def.Status,
	}
	return job
}

func JobToDef(job Job) codec.Definition {
	def := codec.Definition{
		Name:     job.JobName,
		Schedule: job.Schedule,
		Host:     job.Host,
		Commands: job.Commands,
		Env:      job.Env,
		Status:   job.Status,
	}
	return def
}
