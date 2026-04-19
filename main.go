package main

import (
	"net/http"
	"html/template"
	"log"
	"os"
	"database/sql"
	_ "modernc.org/sqlite"
	_ "embed"
)

//go:embed sql/schema.sql
var schema string
var db *sql.DB

func initDB(path string) {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA foreign_keys=ON;")
	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}
}

var templates map[string]*template.Template
func loadTemplates() {
	templates = make(map[string]*template.Template)
	pages := []string{"home", "jobs"}
	for _, page := range pages {
		t := template.Must(template.ParseFiles(
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
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates["home"].ExecuteTemplate(w, "base.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }	
}

type Job struct {
	ID int
	JobName string
	Schedule string
	Host string
	JobStatus string
	JobType string
	Commands string
	Created string
	Updated string
}

func (j *Job) saveJob() error {
	result, err := db.Exec(
		`INSERT 
		INTO jobs (JobName, Schedule, Host, JobStatus, JobType, Commands, Created, Updated) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		j.JobName,
		j.Schedule,
		j.Host,
		j.JobStatus,
		j.JobType,
		j.Commands,
		j.Created,
		j.Updated,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	j.ID = int(id)
	return nil
}
func getAllJobs() ([]Job, error) {
	var jobs []Job
	rows, err := db.Query("SELECT * FROM jobs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var j Job
		err := rows.Scan(
			&j.ID, 
			&j.JobName, 
			&j.Schedule, 
			&j.Host, 
			&j.JobStatus,
			&j.JobType, 
			&j.Commands, 
			&j.Created, 
			&j.Updated,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func getJobs(ids []int) ([]Job, error) {
	var jobs []Job
	for _, id := range ids {
		var j Job
		err := db.QueryRow("SELECT * FROM jobs where id = ?", id).
			Scan(
				&j.ID, 
				&j.JobName, 
				&j.Schedule, 
				&j.Host, 
				&j.JobStatus,
				&j.JobType, 
				&j.Commands, 
				&j.Created, 
				&j.Updated,
			)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	jobs, err := getAllJobs()
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	render(w, "jobs", map[string]any{"Jobs": jobs})
}


func main() {
	dbPath := os.Getenv("RITUAL_DB_PATH")
		if dbPath == "" { 
			dbPath = "./ritual.db" 
		}
	port := os.Getenv("RITUAL_PORT")
		if port == "" {
			port = "8080"
		}
	
	initDB(dbPath)
	defer db.Close()
	loadTemplates()
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("GET /{$}", homeHandler) //Home/Landing Page
	http.HandleFunc("GET /jobs", jobsHandler) //Jobs Page
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}