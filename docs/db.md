# `internal/db`

**Files:** `db/db.go`, `db/job.go`, `db/run.go`, `db/hosts.go` (+ embedded `schema.sql`)

## Purpose

The **only** package that touches SQLite. It owns the connection, the schema, the
`Job` and `Run` models, and all their CRUD. Everything else goes through here for
persistence — keeping `db` "pure" (no HTTP, no bus, no scheduling) is a deliberate
layering rule.

## How it works

### Connection (`db.go`)
`Init()` resolves the path from `$RITUAL_DB_PATH` (default `./ritual.db`), connects
via `sqlx` + `modernc.org/sqlite` (cgo-free) with **WAL mode** and **foreign keys**
on, then executes the embedded `schema.sql` (idempotent `CREATE`s). The connection
is stored in the package global `db.DB` that the query helpers use. `Close()` closes
it.

### Jobs (`job.go`)
The `Job` struct mirrors the Jobs table (id, name, schedule, host, commands, env,
hash, status, timestamps, last/next run). Operations:

- **`CreateJob`** — computes the [hash](codec.md), `INSERT … ON CONFLICT(Hash) DO
  NOTHING`; if nothing was inserted it looks up the colliding job and returns a
  descriptive error (dedup).
- **`UpdateJob`** — full-struct overwrite via `NamedExec` (chosen over per-column
  dynamic SQL because `?` can't bind column names). Bumps `Updated`.
- **`CalcNextRun`** — parses the schedule with robfig, computes the next fire time,
  persists via `UpdateJob`.
- **`GetJob` / `GetJobs` / `GetAllJobs` / `DeleteJob`** — straightforward reads/delete.

### Env serialization: `envMap`
`type envMap map[string]string` implements `driver.Valuer` + `sql.Scanner`, so the
map is stored as sorted `KEY=val\n` text and reconstructed on read — the conversion
lives in one place (`EnvMapToString` / `EnvStringToMap`).

### Runs (`run.go`)
The `Run` struct is one execution record (job id/name, host, start/end, duration,
exit code, logs). `CreateRun()` inserts it. `TimeStamp` is a `time.Time` wrapper
implementing `Valuer`/`Scanner` to format times consistently (`SqlTimeFormat`,
stored UTC).

### Hosts (`hosts.go`)
The `Host` struct mirrors the Hosts table (id, name, address, port, user, key path) —
the connection identity the [SSH runner](cron.md) needs. Only **`GetHost(name)`**
exists today (a single read); there is no create/list/delete yet, so a host can only
be added by raw SQL. Private keys and `known_hosts` are read from the executing user's
`~/.ssh`, not the DB.

### Conversion
`DefToJob` / `JobToDef` bridge to [`codec.Definition`](codec.md) for import/export.

## Status & future

From [TODO.md](../TODO.md):

- **DB path is cwd-relative** — daemon and CLI can open different files; resolve to a
  fixed path.
- **Runs don't update the Job.** `ExecuteJob` writes only a `Runs` row now, so
  `Jobs.LastRun`/`NextRun` are never stamped. `CalcNextRun` is defined here but called
  nowhere — wire it into create/edit so `NextRun` is known before the first run.
- Schedule should be validated (`robfig.ParseStandard`) on create/edit rather than
  failing silently later in `AddFunc`.
- The env-string format is duplicated with [`codec`](codec.md); fold into a shared
  `EnvMap` type.
- Host management CRUD (`CreateHost`/`ListHosts`/`DeleteHost`) is missing — see TODO.
- `schema.sql`'s `Hosts.KeyPath` default contains a literal `{$USER}` that SQLite does
  not expand.
</content>
