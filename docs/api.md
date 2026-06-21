# `internal/api`

**File:** `api/api.go`

## Purpose

The **JSON HTTP edge**. Thin handlers that decode a request, call an
[`ops`](ops.md) operation, and encode the response. All the "marshal at the
boundary, structs in-process" serialization lives here, not in `ops` or
[`db`](db.md). These routes are served on both the unix socket and the TCP listener
(same mux — see [`srv`](srv.md)).

## How it works

Each handler is the same three-step shape:

```go
func createJobHandler(w, r) {
    var request ops.RequestBody
    json.NewDecoder(r.Body).Decode(&request)   // bytes → struct
    response, err := request.CreateJobCall()    // in-process op call
    // ... write status ...
    json.NewEncoder(w).Encode(&response)         // struct → bytes
}
```

Routes are registered onto a caller-owned mux:

```go
func Register(mux *http.ServeMux) {
    mux.HandleFunc("POST /api/jobs/new", createJobHandler)   // create a job via ops
    mux.HandleFunc("POST /api/publish",  publishEventHandler) // forward bus events
}
```

`/api/publish` is the endpoint the [CLI](cmd.md)'s `publishToDaemon` hits to tell the
running daemon "these job IDs changed, reload them."

## Status & future

- This is the machine/JSON namespace (`/api/...`), kept separate from the HTML
  namespace in [`web`](web.md) — representation is chosen by route, not by content
  negotiation.
- `Register(mux)` already takes the mux as a parameter (good), but `srv` still keeps
  the mux as a package global; the design intent is fuller dependency injection
  (passing `ops`/deps in rather than reaching for globals) — see the API design notes
  in the project memory.
- More routes (list/get/update/delete/pause) will land here as the matching `ops`
  verbs are added. A future split of muxes (admin-only ops on the socket, safe subset
  on public TCP) is anticipated for when the TCP listener is exposed beyond localhost.
</content>
