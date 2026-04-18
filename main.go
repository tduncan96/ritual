package main

import (
	"net/http"
	"html/template"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var pool *sql.DB

var templates map[string]*template.Template
func loadTemplates() {
	templates = make(map[string]*template.Template)
	pages := []string{"home"}
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

func main() {
	loadTemplates()
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("GET /{$}", homeHandler) //Main Page
	
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}