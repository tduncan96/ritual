package db

import (
	"database/sql"
	"database/sql/driver"
	_ "embed"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type EnvMap map[string]string

type Job struct {
	ID        int    `db:"ID"`
	JobName   string `db:"JobName" toml:"name"`
	Schedule  string `db:"Schedule" toml:"schedule"`
	Host      string `db:"Host" toml:"host"`
	JobType   string `db:"JobType" toml:"type"`
	Commands  string `db:"Commands" toml:"commands"`
	Env       EnvMap `db:"Env"`
	JobStatus string `db:"JobStatus"`
	Created   string `db:"Created"`
	Updated   string `db:"Updated"`
	LastRun   string `db:"LastRun"`
	NextRun   string `db:"NextRun"`
}

var JobDB *sqlx.DB

//go:embed schema.sql
var schema string

func Init(path string) (*sqlx.DB, error) {
	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	cnxn, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := cnxn.Exec(schema); err != nil {
		return nil, err
	}

	JobDB = cnxn
	return JobDB, nil
}

func Close(db *sqlx.DB) {
	db.Close()
}

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

func (j *Job) CreateJob() (int64, error) {
	result, err := JobDB.Exec(
		`INSERT INTO jobs (JobName, Schedule, Host, Commands, Env) 
		VALUES (?, ?, ?, ?, ?)`,
		j.JobName,
		j.Schedule,
		j.Host,
		j.Commands,
		j.Env,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	j.ID = int(id)
	return id, nil
}

func (j *Job) UpdateJob() (int64, error) {
	result, err := JobDB.NamedExec(
		`UPDATE jobs SET
				JobName  = :JobName,
				Schedule = :Schedule,
				Host     = :Host,
				JobType  = :JobType,
				Commands = :Commands,
				Env      = :Env,
				JobStatus = :JobStatus,
				Updated  = datetime('now')
				WHERE ID = :ID`,
		j,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func DeleteJob(id int) (int64, error) {
	result, err := JobDB.Exec("DELETE FROM jobs WHERE ID = ?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetJob(id int) (Job, error) {
	var job Job
	err := JobDB.Get(&job, "SELECT * FROM Jobs WHERE Id = ?", id)
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
			fmt.Printf("error getting job ID: %v", id)
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func GetAllJobs() ([]Job, error) {
	var jobs []Job
	var ids []int
	err := JobDB.Select(&ids, "SELECT Id FROM Jobs")
	if err != nil {
		return []Job{}, err
	}
	jobs, err = GetJobs(ids)
	if err != nil {
		return []Job{}, err
	}

	return jobs, nil
}
