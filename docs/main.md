# `main` (root package)

**File:** `main.go`

## Purpose

The process entrypoint. It's deliberately tiny: open the database, then hand control
to the Cobra command tree.

## How it works

```go
func main() {
    database, err := db.Init()   // open/connect SQLite, apply schema
    if err != nil { os.Exit(1) }
    defer db.Close(database)

    if err := cmd.Execute(); err != nil { os.Exit(1) }
}
```

Two things to notice:

- **The DB is opened on *every* invocation** — both `ritual serve` (the daemon) and
  every short-lived CLI command (`ritual import`, `ritual run`, …). The CLI process
  therefore holds its own `*sqlx.DB` handle even when it's only going to forward a
  request to the daemon over the socket. (WAL mode + `foreign_keys` are set in
  [`db.Init`](db.md).)
- **`defer db.Close`** runs when `main` returns. For the daemon that means the DB is
  closed only after the signal-driven shutdown in [`cmd serve`](cmd.md) unblocks.

## Status & future

- The DB path comes from `$RITUAL_DB_PATH`, defaulting to `./ritual.db` (cwd-relative).
  This is a known footgun — the daemon and a CLI launched from a different directory
  can open *different* database files. Resolving this to a fixed path is tracked in
  [TODO.md](../TODO.md) (Bugs).
- Opening SQLite unconditionally is fine today, but once the CLI reliably routes
  through the daemon it may not need its own handle except for the daemon-down
  fallback path.
</content>
