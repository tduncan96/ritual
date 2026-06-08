<p align="center">
  <img src="internal/web/static/images/logo_no_bg.png" alt="Ritual" width="150">
</p>

<p align="center"><b>Ritual</b><br>A self-hosted job scheduler that gives cron a memory and a face.</p>

## Overview

Ritual is a small Go service for defining, scheduling, running, and reviewing
recurring jobs. Plain crontab schedules a command and forgets it. Ritual keeps
the schedule, runs the command, records every run (exit code, output, timing),
and tracks when each job last ran and runs next. You drive it from a CLI or a
built-in web UI, and you can pull in jobs you already have by importing existing
crontabs or TOML files.

## Features

- Create and schedule jobs from the CLI or the web UI.
- Standard cron expressions plus shortcuts like `@every` and `@hourly`.
- Real execution through a shell, with exit code and combined output captured.
- Run history stored per job, with last-run and next-run timestamps.
- Import jobs from a local crontab or from TOML definitions.
- Run any job on demand without waiting for its schedule.

## Build

Requires Go 1.26 or newer.

```
go build -o ritual .
```

## Usage

```
# start the scheduler and the web UI (http://localhost:1771)
ritual start

# create a job: name, schedule, host, command, [env file]
ritual create "nightly-backup" "0 2 * * *" local "/usr/local/bin/backup.sh"

# run a job now by id
ritual run 1

# import jobs
ritual import cron localhost
ritual import toml ./jobs/backup.toml
```

## Configuration

| Variable           | Default         | Purpose                          |
| ------------------ | --------------- | -------------------------------- |
| `RITUAL_DB_PATH`   | `./ritual.db`   | SQLite database location         |
| `RITUAL_TOML_DUMP` | `./toml-dump/`  | Directory for TOML job files     |

The web UI currently serves on port 1771.

## Status

Ritual is in active development. The core loop (define, schedule, run, log) works
today. Remote execution over SSH, authentication, a run-history UI, and a TUI are
on the roadmap.
</content>
