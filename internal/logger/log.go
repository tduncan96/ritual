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
	return Logger{logger: log.With().Str("component", component).Logger()}
}

func (l Logger) Debug() Event { return Event{Event: l.logger.Debug()} }
func (l Logger) Info() Event  { return Event{Event: l.logger.Info()} }
func (l Logger) Warn() Event  { return Event{Event: l.logger.Warn()} }
func (l Logger) Error() Event { return Event{Event: l.logger.Error()} }
func (l Logger) Fatal() Event { return Event{Event: l.logger.Fatal()} }

type Event struct {
	*zlog.Event
}

func (e Event) Err(err error) Event {
	e.Event.Err(err)
	return e
}

func (e Event) Job(v Verb, j db.Job) Event {
	e.Event.
		Str("verb", string(v)).
		Int64("job_id", j.JobId).
		Str("job_name", j.JobName)
	if j.Host != nil {
		e.Event.Str("host", *j.Host)
	}
	return e
}

func (e Event) Run(r db.Run) Event {
	e.Event.
		Int64("run_id", r.RunId).
		Dur("duration", time.Duration(r.Duration)).
		Int64("exit_code", r.ExitCode)
	return e
}

func init() {
	zlog.TimeFieldFormat = "2006-01-02 15:04:05"
	zlog.TimestampFunc = func() time.Time { return time.Now().UTC() }

	log.Logger = log.Logger.With().Str("service", "ritual").Logger()
}
