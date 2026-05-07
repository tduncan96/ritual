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
	
	db.InitDB(dbPath)
	db.DbDontClose()
	
	web.Start(port)
}