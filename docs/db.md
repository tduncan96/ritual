# `internal/db`

**Files:** `db/db.go`, `db/job-db.go`, `db/run-db.go` (+ embedded `schema.sql`)

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

### Jobs (`job-db.go`)
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

### Runs (`run-db.go`)
The `Run` struct is one execution record (job id/name, host, start/end, duration,
exit code, logs). `CreateRun()` inserts it. `TimeStamp` is a `time.Time` wrapper
implementing `Valuer`/`Scanner` to format times consistently (`SqlTimeFormat`,
stored UTC).

### Conversion
`DefToJob` / `JobToDef` bridge to [`codec.Definition`](codec.md) for import/export.

## Status & future

From [TODO.md](../TODO.md):

- **DB path is cwd-relative** — daemon and CLI can open different files; resolve to a
  fixed path.
- **`Updated` is bumped on every run**, not just on edits (because the runner's
  bookkeeping calls `UpdateJob`). Runs shouldn't count as edits.
- `NextRun` should be computed on create/edit, not only after the first run.
- Schedule should be validated (`robfig.ParseStandard`) on create/edit rather than
  failing silently later in `AddFunc`.
- The env-string format is duplicated with [`codec`](codec.md); fold into a shared
  `EnvMap` type.
</content>
