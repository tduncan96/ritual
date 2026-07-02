# `internal/run`

**File:** `run/runner.go` (execution)

## Purpose

Executes a job's command — locally or over SSH — and records the result as a
[`db.Run`](db.md). This is the "do the work" layer that both the [scheduler](cron.md)
(via its `AddFunc` closure) and the `ritual run <id>` CLI reach.

```go
type Runner struct {
    Job    db.Job
    Client *ssh.Client   // nil ⇒ run locally
}
```

## How a run happens

`ExecuteJob` owns the bookkeeping; `ResolveTarget` and `RunCommand` own the difference
between local and remote:

1. **`ResolveTarget`** decides where to run by `Job.Host`:
   - `""` → error ("invalid host");
   - `"localhost"` → run locally (`Client` stays nil);
   - anything else → look the host up via [`db.GetHost`](db.md), read its private key
     (`~` expanded to `$HOME`), build an `ssh.ClientConfig` (key auth + `known_hosts`
     checking, 10s dial timeout), and dial — storing the live `*ssh.Client` on the `Runner`.
2. record start time, seed a `db.Run`;
3. **`RunCommand`** prepends `export KEY='val'` lines for the job's `Env`, then runs the
   command: `sh -c` + `CombinedOutput` locally, or an `ssh.Session` + `CombinedOutput`
   remotely. A non-zero exit is **recorded, not treated as a failure** (exit code pulled
   from `*exec.ExitError` / `*ssh.ExitError`).
4. **`CalcNextRun`** stamps `LastRun`/`NextRun` back onto the Job;
5. write the `db.Run` (timing, duration, exit code, logs). Errors from the command,
   `CalcNextRun`, and the write are joined via `errors.Join`.

## Status & future

See [TODO.md](../TODO.md) for the live queue.

- **`ResolveTarget` derefs `*Job.Host` with no nil case** (`runner.go:65`). `Job.Host`
  is a `*string`, and `ON DELETE SET NULL` can produce a NULL host, so a nil host panics
  the runner here (and in `GetHash` on the create path). The fix is a shared
  `isLocal(host)` — nil / `""` / `"localhost"` ⇒ local — used by both execution and
  crontab import so they agree on what "local" means. This is the same gap that keeps
  local job *creation* failing the FK (see High in TODO.md).
- No per-job timeout (`exec.CommandContext`), overlap guard, or `recover()` around runs
  yet.
