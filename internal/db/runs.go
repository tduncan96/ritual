package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

// Run
type Run struct {
	RunId     int64     `db:"RunId"`
	JobId     *int64    `db:"JobId"`
	JobName   string    `db:"JobName"`
	Host      *string   `db:"Host"`
	StartTime TimeStamp `db:"StartTime"`
	EndTime   TimeStamp `db:"EndTime"`
	Duration  int64     `db:"Duration"`
	ExitCode  int64     `db:"ExitCode"`
	Logs      string    `db:"Logs"`
}

func (run *Run) CreateRun() (int64, error) {
	result, err := DB.NamedExec(
		`INSERT INTO Runs (JobId, JobName, Host, StartTime, EndTime, Duration, ExitCode, Logs) 
		VALUES (:JobId, :JobName, :Host, :StartTime, :EndTime, :Duration, :ExitCode, :Logs)`,
		run,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// TimeStamp
type TimeStamp time.Time

var _ driver.Valuer = TimeStamp{}
var _ sql.Scanner = (*TimeStamp)(nil)

func (ts TimeStamp) Value() (driver.Value, error) {
	return time.Time(ts).UTC().Format(SqlTimeFormat), nil
}

func (ts *TimeStamp) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("TimeStamp.Scan: expected string, got %T", src)
	}
	parsed, err := time.Parse(SqlTimeFormat, s)
	if err != nil {
		return err
	}
	*ts = TimeStamp(parsed)
	return nil
}
