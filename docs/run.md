# `internal/run`

**Files:** `run/runner.go`, `run/local.go`, `run/remote.go`

## Purpose

Actually executes a job's command and records the result as a [`db.Run`](db.md).
This is the "do the work" layer the scheduler and the `ritual run` CLI both call.

## How it works

### The interface
```go
type Runner interface {
    ExecuteJob(job db.Job) error
}
```

Callers pick an implementation by host (see the dispatch in [`cron`](cron.md) and
[`cmd`](cmd.md)): `"localhost"` → `LocalRunner`, anything else → `RemoteRunner`.

### `LocalRunner` (real)
The reference implementation of "what an execution is":

1. record start time;
2. `exec.Command("sh", "-c", job.Commands)` with `os.Environ()` + the job's `Env`;
3. `CombinedOutput()` — capture stdout+stderr together;
4. extract the exit code from `*exec.ExitError` (a non-zero exit is **recorded, not
   treated as a failure**);
5. build a `db.Run` (timing, duration, exit code, logs) and `CreateRun()`;
6. stamp `LastRun` and `CalcNextRun()`.

### `RemoteRunner` (stub)
Currently `ExecuteJob` returns `nil` and does nothing — the remote/SSH path is not
built yet.

## Status & future

This package is the focus of the SSH work. The full design is in
[../EXPLAIN.md](../EXPLAIN.md); in short:

- **Collapse to one `ExecuteJob` + a small `Target` interface.** Instead of two full
  `Runner`s duplicating the bookkeeping, keep one function that owns steps 1 and 5–6,
  and a 2-method `Target { Run(cmd, env) (out, exitCode, err); Close() error }` for
  the only part that differs (step 2–4). A `connect(job)` helper decides local vs
  remote once and returns a ready-to-run `Target`.
- **SSH via `golang.org/x/crypto/ssh`** (not stdlib): dial TCP → session →
  `CombinedOutput` → exit code from `*ssh.ExitError`. It mirrors the local runner
  almost call-for-call.
- **Topology in the DB, secrets in `~/.ssh`** — a `hosts` table (address, port, user,
  nullable key path) for connection identity; private keys + `known_hosts` read from
  the executing user's `~/.ssh`.
- **Fix `isLocal`** (TODO bug): the `"localhost"`-literal check misses jobs whose host
  is the machine's real hostname; centralize the decision in `connect`.

Per-job execution timeout (`exec.CommandContext`), an overlap guard, and a
`recover()` around runs are also tracked in [TODO.md](../TODO.md).
</content>
