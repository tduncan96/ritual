package main

import (
	"fmt"
	"os"
	"ritual/cmd"
	"ritual/internal/db"
)

func main() {
	database, err := db.Init()
	if err != nil {
		fmt.Print("Database initialization error. Exiting ...")
		os.Exit(1)
	}
	defer db.Close(database)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
