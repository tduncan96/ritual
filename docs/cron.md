# `internal/cron`

**File:** `cron/cron.go`

## Purpose

Wraps `robfig/cron/v3` into Ritual's live scheduler and keeps it in sync with the
database. It's the bridge between "rows in the Jobs table" and "things that actually
fire on a schedule."

## How it works

### `CronRunner`
```go
type CronRunner struct {
    Cron   *robfig.Cron
    Lookup map[int64]robfig.EntryID   // JobId → scheduler entry
}
```

The `Lookup` map is the key idea: robfig hands back an `EntryID` when you add a job,
and you need it to later remove/replace that entry. Keeping `JobId → EntryID` is what
makes edit/delete reach the running scheduler.

### Building & syncing
- **`MakeRunner`** — creates the cron, loads all jobs from [`db`](db.md), adds them.
- **`AddJobs`** — for each enabled job, registers an `AddFunc(schedule, closure)`.
  The closure picks a [`run.Runner`](run.md) by host and calls `ExecuteJob`, logging
  errors. The returned `EntryID` is stored in `Lookup`.
- **`UpdateRunner(ids)`** — reloads those jobs from the DB, removes their old entries,
  re-adds them. This is what an edit/create event triggers.
- **`RemoveRunnerJob(ids)`** — removes entries and drops them from `Lookup`.

The scheduler is driven by events: [`bus.CronSubscription`](bus.md) calls
`UpdateRunner` / `RemoveRunnerJob` / `Cron.Start` / `Cron.Stop` in response to bus
events published by [`ops`](ops.md).

```mermaid
flowchart LR
    OPS[ops mutation] -->|publish Database event| BUS[bus]
    BUS -->|UpdateRunner ids| CR[CronRunner]
    CR -->|AddFunc / Remove| ROBFIG[robfig clock]
    ROBFIG -->|time up| RUN[run.ExecuteJob]
```

## Status & future

From [TODO.md](../TODO.md):

- **`UpdateRunner` does `Stop()`/`Start()`** around the swap, which makes robfig
  recompute `@every` next-runs from "now" → relative schedules drift on every edit.
  Should use live `AddFunc`/`Remove` without stopping the whole cron.
- **`isLocal` dispatch** is a literal `job.Host == "localhost"` check (`cron.go:39`),
  so imported jobs (host = real hostname) never run locally. Needs a shared
  `isLocal(host)` helper — see [run.md](run.md)/[EXPLAIN.md](../EXPLAIN.md).
- Logging nits: a `fmt.Sprintf` with no verbs but extra args, and a duplicated "cron
  runner jobs updated" line (also logged in [`bus`](bus.md)).
- Closures capture `job` by value, which is correct (an edit reschedules with a fresh
  snapshot), but be careful preserving that as this code evolves.
</content>
