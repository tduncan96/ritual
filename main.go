package main

import (
	"os"
	"ritual/internal/db"
	"ritual/web"
)

func main() {
	dbPath := os.Getenv("RITUAL_DB_PATH")
		if dbPath == "" { 
			dbPath = "./ritual.db" 
		}
	port := os.Getenv("RITUAL_PORT")
		if port == "" {
			port = "8080"
		}
	
	database, _ :=db.InitDB(dbPath)
	defer db.Close(database)
	
	s := &web.Server{DB: database}

	s.Start(port)
}