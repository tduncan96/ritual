package db

import (
	"log"
	"database/sql"
	_ "modernc.org/sqlite"
	_ "embed"
)

type Job struct {
	ID int
	JobName string
	Schedule string
	Host string
	JobStatus string
	JobType string
	Commands string
	Created string
	Updated string
	LastRun string
	NextRun string
}

//go:embed schema.sql
var schema string
var db *sql.DB

func InitDB(path string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA foreign_keys=ON;")
	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}
	return db, nil
}

func DbDontClose(db *sql.DB) {
	defer db.Close(db)
}

func (j *Job) CreateJob() (int64, error) {
	result, err := db.Exec(
		`INSERT INTO jobs (JobName, Schedule, Host, JobStatus, JobType, Commands, LastRun, NextRun) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		j.JobName,
		j.Schedule,
		j.Host,
		j.JobStatus,
		j.JobType,
		j.Commands,
		j.LastRun,
		j.NextRun,
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

func DeleteJob(id int) (int64, error) {
	result, err := db.Exec("DELETE FROM jobs WHERE ID = ?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetAllJobs() ([]Job, error) {
	var jobs []Job
	rows, err := db.Query("SELECT * FROM jobs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var j Job
		err := rows.Scan(
			&j.ID, 
			&j.JobName, 
			&j.Schedule, 
			&j.Host, 
			&j.JobStatus,
			&j.JobType, 
			&j.Commands, 
			&j.Created, 
			&j.Updated,
			&j.LastRun,
			&j.NextRun,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func GetJobs(ids []int) ([]Job, error) {
	var jobs []Job
	for _, id := range ids {
		var j Job
		err := db.QueryRow("SELECT * FROM jobs where id = ?", id).
			Scan(
				&j.ID, 
				&j.JobName, 
				&j.Schedule, 
				&j.Host, 
				&j.JobStatus,
				&j.JobType, 
				&j.Commands, 
				&j.Created, 
				&j.Updated,
			)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

