# `internal/web`

**File:** `web/web.go` (+ embedded `templates/*.gohtml`, `static/`)

## Purpose

The **HTML web UI** — server-rendered pages for browsing and managing jobs. Templates
and static assets are embedded into the binary, so the daemon ships as a single file.
Served on the TCP listener (`:1771`) via the shared mux.

## How it works

### Templates
- `//go:embed templates/*.gohtml` and `//go:embed static` bake the assets in. (The
  embed must stay in this package — `go:embed` paths are relative to the source file.)
- **`LoadTemplates`** parses each page template together with the shared
  `base.gohtml` layout into a `map[page]*template.Template`. Called once at startup by
  [`srv.WebServe`](srv.md).
- **`render`** executes a named template against the `base` layout into a buffer, then
  writes it out (buffering first so a mid-render error doesn't emit a half page).

### Handlers & routes
```go
func Register(mux *http.ServeMux) {
    mux.Handle("GET /static/", http.FileServer(http.FS(staticFS)))
    mux.HandleFunc("GET /{$}",        homeHandler)        // landing
    mux.HandleFunc("GET /jobs",       jobsHandler)        // list
    mux.HandleFunc("GET /jobs/new",   jobFormHandler)     // create form
    mux.HandleFunc("POST /jobs/new",  createJobHandler)   // submit
    mux.HandleFunc("GET /jobs/{id}",  jobHandler)         // detail
    mux.HandleFunc("DELETE /jobs/{id}", deleteJobHandler) // delete
}
```

Handlers read/write jobs through [`db`](db.md) directly (e.g. `createJobHandler`
builds a `db.Job` from the form and calls `CreateJob`).

## Status & future

- **The web handlers currently call [`db`](db.md) directly, bypassing
  [`ops`](ops.md) and the [bus](bus.md).** That means a job created through the web UI
  does *not* publish a reload event, so the live scheduler won't pick it up until
  restart. Routing web mutations through `ops` (same as the planned CLI path) is the
  intended fix.
- A rewrite is anticipated (the `LoadTemplates` comment in `srv` flags it). The
  roadmap also wants a **Bubble Tea TUI** that mirrors these same views over the same
  [`db`](db.md)/[`ops`](ops.md) core — which is the reason to keep logic *out* of this
  package.
- Minor: `render` uses `%w` on a non-error string in one `fmt.Errorf` (go vet) — TODO.
</content>
