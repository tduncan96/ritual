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

	"github.com/jmoiron/sqlx"
)

type Server struct {
	DB *sqlx.DB
}

//go:embed templates/*.gohtml
var templateFS embed.FS
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

func (s *Server) render(w http.ResponseWriter, templateName string, data any) {
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

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	s.render(w, "home", nil)
}

func (s *Server) createJobHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) deleteJobHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) jobsHandler(w http.ResponseWriter, r *http.Request) {
	jobs, err := db.GetAllJobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "jobs", map[string]any{"Jobs": jobs})
}

func (s *Server) jobFormHandler(w http.ResponseWriter, r *http.Request) {
	s.render(w, "job_form", nil)
}

func (s *Server) jobHandler(w http.ResponseWriter, r *http.Request) {
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
	s.render(w, "job", job)
}

//go:embed static
var staticFS embed.FS

func (s *Server) Start(port string) {
	loadTemplates()
	http.Handle("GET /static/", http.FileServer(http.FS(staticFS)))

	http.HandleFunc("GET /{$}", s.homeHandler)               // Home Landing Page
	http.HandleFunc("GET /jobs", s.jobsHandler)              // Jobs Page
	http.HandleFunc("GET /jobs/new", s.jobFormHandler)       // New Job Creation Form
	http.HandleFunc("POST /jobs/new", s.createJobHandler)    // Submit New Job Form
	http.HandleFunc("GET /jobs/{id}", s.jobHandler)          // Individual Job page
	http.HandleFunc("DELETE /jobs/{id}", s.deleteJobHandler) // Delete Job

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
