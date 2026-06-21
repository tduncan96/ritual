# `codec`

**Files:** `codec/codec.go`, `codec/cron-codec.go`, `codec/toml-codec.go`, `codec/yaml-codec.go`

## Purpose

Translates between job **definitions** and the file formats users keep them in
(crontab, TOML, YAML), in both directions. Also owns the **content hash** used to
deduplicate jobs. This is the import/export layer; it knows nothing about SQLite or
HTTP.

## How it works

### The shared shape: `Definition`
A transport-/storage-neutral struct with struct tags for every format:

```go
type Definition struct {
    Name, Schedule, Host, Commands string
    Env    map[string]string
    Hash   string   // never serialized
    Status bool
}
```

`db.DefToJob` / `db.JobToDef` ([db](db.md)) convert between this and the persisted
`db.Job`.

### The `Codec` interface + registry
```go
type Codec interface {
    Marshal([]Definition) ([]byte, error)    // → file bytes
    Unmarshal([]byte) ([]Definition, error)  // → structs
}
var Codecs = map[string]Codec{"cron": …, "toml": …, "yaml": …}
```

The CLI looks a codec up by file extension (`Codecs[fileType]`), so adding a format
is "implement the interface + add one map entry."

- **`TOMLCodec` / `YAMLCodec`** are thin wrappers over `BurntSushi/toml` and
  `goccy/go-yaml`, nesting definitions under a `Rituals` list.
- **`CronCodec`** is the substantial one — it parses/writes actual crontab syntax:
  - *Unmarshal* scans line by line: `KEY=val` lines accumulate into the env map;
    `@every`, other `@macros`, and 5-field schedules are split into schedule+command;
    `## name:` comments name the next job; a `##`-prefixed schedule line marks the job
    paused (`Status=false`). Each schedule is validated with `robfig.ParseStandard`.
    Host defaults to the local machine's `hostname`. Names default to
    `host_crontab_<hash>` when absent.
  - *Marshal* emits `## name:` headers, `KEY=val` env lines, and the schedule+command
    (prefixed with `## ` when paused).

### Hashing: `GetHash`
SHA-256 over `host`, `schedule`, `commands`, and sorted `KEY=val` env lines
(NUL-separated). The result is a job's identity — `db.CreateJob` stores it with a
unique constraint, so re-importing the same definition is a no-op rather than a
duplicate.

## Status & future

- The older memory mentions a JSON codec; only **cron/toml/yaml** exist in the code
  today.
- Known `CronCodec` bugs in [TODO.md](../TODO.md): Marshal emits stray blank lines
  and merges env into the schedule line; `@reboot` lines are silently dropped
  (should warn and skip).
- The env-serialization format (sorted `KEY=val\n`) is duplicated between this
  package's hashing and [`db`](db.md)'s `envMap`; a shared `EnvMap` type is the
  intended consolidation.
</content>
