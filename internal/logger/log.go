package logger

import (
	"time"

	zlog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"ritual/internal/db"
)

type Verb string

const (
	Execute Verb = "Execute"
	Create  Verb = "Create"
	Update  Verb = "Update"
	Delete  Verb = "Delete"
	Serve   Verb = "Serve"
	Stop    Verb = "Stop"
)

type Logger struct {
	logger zlog.Logger
}

func For(component string) Logger {
	logger := log.With().
		Str("service", "ritual").
		Str("component", component).
		Logger()
	return Logger{logger: logger}
}

func (l Logger) Info() Event  { return Event{Event: l.logger.Info()} }
func (l Logger) Warn() Event  { return Event{Event: l.logger.Warn()} }
func (l Logger) Error() Event { return Event{Event: l.logger.Error()} }

type Event struct {
	*zlog.Event
}

func (e Event) Err(err error) Event {
	e.Event.Err(err)
	return e
}

func (e Event) Job(v Verb, j db.Job) Event {
	e.Event.
		Str("Verb", string(v)).
		Int64("JobId", j.JobId).
		Str("JobName", j.JobName).
		Str("Host", *j.Host)
	return e
}

func (e Event) Run(r db.Run) Event {
	e.Event.
		Int64("RunId", r.RunId).
		Dur("Duration", time.Duration(r.Duration)).
		Int64("ExitCode", r.ExitCode)
	return e
}

func init() {
	zlog.TimeFieldFormat = "2006-01-02 15:04:05"
	zlog.TimestampFunc = func() time.Time { return time.Now().UTC() }
}
