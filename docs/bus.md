# `bus`

**File:** `bus/bus.go` (top-level package)

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
- **`Publish(events…)`** — takes the lock only to grab the topic's subscriber slice,
  unlocks, then sends each event to those channels.
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

Fixed since the last audit: a bad payload no longer kills the consumer (`continue`, not
`return`); `Unsubscribe` writes the reslice back into the map; and `Publish` no longer
sends under the lock (it snapshots the slice, unlocks, then sends), so a slow consumer
can only stall that one publish, not deadlock the whole bus.

Remaining, in [TODO.md](../TODO.md):

- **`Publish`'s snapshot is shallow.** `subs := bus.subscribers[topic]` copies the slice
  header but shares the backing array with the map, so a concurrent `Unsubscribe`
  (`slices.Delete`) or `Subscribe` (`append`) races the `range`. Fix: `slices.Clone`.
- A full 16-buffer still *blocks* that publisher (backpressure, not deadlock). If a dead
  consumer should never stall publishers, add a non-blocking send — but that's a
  drop-policy choice; today the intent is *do not* drop events.

Design intent (from the project memory): only user-initiated mutations publish; a
runner's per-run bookkeeping should write straight to [`db`](db.md) silently, never
through the bus, to avoid a reschedule-on-every-run feedback loop. (Today the runner
records nothing back onto the Job at all — see [cron.md](cron.md)/[TODO.md](../TODO.md).)
</content>
