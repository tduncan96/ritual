package web

import (
	"log"
	"embed"
	"net/http"
	"html/template"
	"strconv"
	"ritual/internal/db"
)

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

func render(w http.ResponseWriter, page string, data any) {
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates["home"].ExecuteTemplate(w, "base.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }	
}

func createJobHandler(w http.ResponseWriter, r *http.Request) {
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
	err := templates["job_form"].ExecuteTemplate(w, "base.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Start(port string) {
	loadTemplates()
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("GET /{$}", homeHandler) //Home Landing Page

	http.HandleFunc("GET /jobs", jobsHandler) //Jobs Page
	http.HandleFunc("GET /jobs/new", jobFormHandler) // New Job Creation Form
	http.HandleFunc("POST /jobs/new", createJobHandler) // Submit New Job Form
	http.HandleFunc("DELETE /jobs/{id}", deleteJobHandler) // Delete Job
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}