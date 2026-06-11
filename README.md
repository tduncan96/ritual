<p align="center">
  <img src="internal/web/static/images/logo_no_bg.png" alt="Ritual" width="150">
</p>

<p align="center"><b>Ritual</b><br>A self-hosted job scheduler.</p>

## About

A small Go service for scheduling recurring jobs. It keeps the schedule, runs the
command, and records every run (exit code, output, timing, last/next run). Driven
from a CLI, with a web UI in progress. Jobs can be imported from a local crontab, 
YAML, JSON, or TOML.

## Build

Requires Go 1.26+.

```
go build -o ritual .
```

## Usage

```
# start the scheduler + web UI (http://localhost:1771)
ritual start

# create a job: name, schedule, host, command, [env file]
ritual create "nightly-backup" "0 2 * * *" local "/usr/local/bin/backup.sh"

# run a job now by id
ritual run 1

# import jobs
ritual import ./jobs/backup.toml
ritual import ./jobs/
ritual import --crontab

# export jobs to files (no ids = all jobs)
ritual export toml
ritual export toml 1 2
```

## Configuration

| Variable           | Default       | Purpose                        |
| ------------------ | ------------- | ------------------------------ |
| `RITUAL_DB_PATH`   | `./ritual.db` | SQLite database location       |
| `RITUAL_CRON_PATH` | —             | Default directory for `import` |
