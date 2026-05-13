package web

import (
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"
	"ritual/internal/db"
	"strconv"
)

type Server struct {
	DB *sql.DB
}

//go:embed templates/*.html
var templateFS embed.FS

var templates map[string]*template.Template

func loadTemplates() {
	templates = make(map[string]*template.Template)
	pages := []string{"home", "jobs", "job_form"}
	for _, page := range pages {
		t := template.Must(template.ParseFS(
			templateFS,
			"templates/base.html",
			"templates/"+page+".html",
		))
	templates[page] = t
	}
}

func (s *Server) render(w http.ResponseWriter, page string, data any) {
	t, ok := templates[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates["home"].ExecuteTemplate(w, "base.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }	
}

func (s *Server) createJobHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	j := db.Job{
		JobName: r.FormValue("job_name"),
		Schedule: r.FormValue("schedule"),
		Host: r.FormValue("host"),
		JobStatus: "Active",
		JobType: r.FormValue("job_type"),
		Commands: r.FormValue("command"),
		LastRun: "Never",
		NextRun: "Fuck You",
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

func (s * Server) jobFormHandler(w http.ResponseWriter, r *http.Request) {
	err := templates["job_form"].ExecuteTemplate(w, "base.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Start(port string) {
	loadTemplates()
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("GET /{$}", s.homeHandler) //Home Landing Page

	http.HandleFunc("GET /jobs", s.jobsHandler) //Jobs Page
	http.HandleFunc("GET /jobs/new", s.jobFormHandler) // New Job Creation Form
	http.HandleFunc("POST /jobs/new", s.createJobHandler) // Submit New Job Form
	http.HandleFunc("DELETE /jobs/{id}", s.deleteJobHandler) // Delete Job
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}