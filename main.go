package main

import (
	"net/http"
	"html/template"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var pool *sql.DB

func handler(w http.ResponseWriter, r *http.Request) {
	
	// Main Page
	t, err := template.New("home.html").ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", handler) //Main Page
	
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}