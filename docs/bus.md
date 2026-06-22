# `internal/bus`

**File:** `bus/bus.go`

## Purpose

An in-process publish/subscribe event bus. It decouples *mutations* (someone created
or changed a job) from *reactions* (the live scheduler needs to reload). Today it has
exactly one consumer — the scheduler — but it's built to grow more (e.g. SSE live
updates for the web UI).

## How it works

### Topics, methods, events
```go
type SubList int   // topic:  LifeCycle, Database
type Method int    // verb:   GET, POST, PUT, DELETE
type Event struct {
    SubList SubList
    Method  Method
    Payload []byte   // opaque — JSON, decoded by the consumer
}
```

The `Payload` is intentionally opaque bytes (e.g. a JSON array of job IDs), keeping
the bus ignorant of what flows through it.

### The bus
A simple **mutex-guarded** map of topic → subscriber channels (not a goroutine
broker):

```go
type EventBus struct {
    mu          sync.Mutex
    subscribers map[SubList][]chan Event
}
var GlobalBus *EventBus   // created by MakeBus() in `serve`
```

- **`Subscribe(topics…)`** — makes a buffered channel (cap 16), registers it under
  each topic, returns the receive end.
- **`Publish(events…)`** — under the lock, sends each event to every channel
  subscribed to its topic.
- **`Unsubscribe(ch, topics…)`** — removes the channel from those topics.

### The consumer: `CronSubscription`
The bridge from bus → [scheduler](cron.md). It subscribes to `LifeCycle` + `Database`
and loops over events:

- `LifeCycle` + `PUT`/`DELETE` → `Cron.Start()` / `Cron.Stop()`.
- `Database` → unmarshal the payload into `[]int64` job IDs, then `POST` →
  `UpdateRunner(ids)`, `DELETE` → `RemoveRunnerJob(ids)`.

Producers are [`ops`](ops.md) (after a successful DB write) and the
[CLI](cmd.md) (via `/api/publish`).

## Status & future

Known issues in [TODO.md](../TODO.md) — read before relying on this:

- **A bad payload kills the consumer permanently:** `CronSubscription` does `return`
  (not `continue`) on an unmarshal error, ending the loop forever.
- **`Unsubscribe` doesn't actually remove the channel:** the reslice isn't written
  back into the map, leaving a `nil` entry that makes `Publish` block forever sending
  to it (staticcheck SA4006).
- **`Publish` sends under the lock**, so if a slow consumer fills its 16-buffer the
  publisher stalls (backpressure, not deadlock). Intended fix is a larger/unbounded
  buffer — *do not* drop events.

Design intent (from the project memory): only user-initiated mutations publish; a
runner's per-run bookkeeping should write straight to [`db`](db.md) silently, never
through the bus, to avoid a reschedule-on-every-run feedback loop. (Today the runner
records nothing back onto the Job at all — see [cron.md](cron.md)/[TODO.md](../TODO.md).)
</content>
