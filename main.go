package main

import (
	"os"
	"fmt"
	"ritual/internal/db"
	// "ritual/internal/web"
	"ritual/cmd"
)

func main() {
	dbPath := os.Getenv("RITUAL_DB_PATH")
		if dbPath == "" { 
			dbPath = "./ritual.db" 
		}

	database, err := db.InitDB(dbPath)
	if err != nil {
		fmt.Print("Database initialization error. Exiting ...")
		os.Exit(1)
	}
	defer db.Close(database)

	cmd.Database = database
	
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}