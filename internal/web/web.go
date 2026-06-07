package web

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"ritual/internal/db"
	"strconv"
	"strings"
	"time"
)

var templates map[string]*template.Template

func loadTemplates() {
	templates = make(map[string]*template.Template)

	entries, err := fs.ReadDir(templateFS, "templates")
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "base.gohtml" {
			continue
		}
		page := strings.TrimSuffix(name, ".gohtml")
		t := template.Must(template.ParseFS(
			templateFS,
			"templates/base.gohtml",
			"templates/"+name,
		))
		templates[page] = t
	}
}

func render(w http.ResponseWriter, templateName string, data any) {
	t, ok := templates[templateName]
	if !ok {
		http.Error(w, "template not found: "+templateName, http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base.gohtml", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	render(w, "home", nil)
}

func createJobHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	j := db.Job{
		JobName:   r.FormValue("job_name"),
		Schedule:  r.FormValue("schedule"),
		Host:      r.FormValue("host"),
		JobStatus: "Active",
		JobType:   r.FormValue("job_type"),
		Commands:  r.FormValue("command"),
	}

	_, err := j.CreateJob()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteJobHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		_, err := db.DeleteJob(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	jobs, err := db.GetAllJobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, "jobs", map[string]any{"Jobs": jobs})
}

func jobFormHandler(w http.ResponseWriter, r *http.Request) {
	render(w, "job_form", nil)
}

func jobHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	job, err := db.GetJob(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, "job", job)
}

//go:embed static
var staticFS embed.FS

//go:embed templates/*.gohtml
var templateFS embed.FS

func Start() {

	loadTemplates()
	http.Handle("GET /static/", http.FileServer(http.FS(staticFS)))

	http.HandleFunc("GET /{$}", homeHandler)               // Home Landing Page
	http.HandleFunc("GET /jobs", jobsHandler)              // Jobs Page
	http.HandleFunc("GET /jobs/new", jobFormHandler)       // New Job Creation Form
	http.HandleFunc("POST /jobs/new", createJobHandler)    // Submit New Job Form
	http.HandleFunc("GET /jobs/{id}", jobHandler)          // Individual Job page
	http.HandleFunc("DELETE /jobs/{id}", deleteJobHandler) // Delete Job

	srv := &http.Server{
		Addr:         ":1771",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
