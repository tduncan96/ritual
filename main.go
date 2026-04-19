package main

import (
	"net/http"
	"html/template"
	"log"
	"os"
	"database/sql"
	_ "modernc.org/sqlite"
)

var pool *sql.DB

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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates["home"].ExecuteTemplate(w, "base.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }	
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	err := templates["jobs"].ExecuteTemplate(w, "base.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	
	loadTemplates()
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("GET /{$}", homeHandler) //Home/Landing Page
	http.HandleFunc("GET /jobs", jobsHandler) //Jobs Page
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}