package db

import (
	_ "embed"
	"os"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const SqlTimeFormat = "2006-01-02 15:04:05"

var DB *sqlx.DB

//go:embed schema.sql
var schema string

func Init() (*sqlx.DB, error) {
	path := os.Getenv("RITUAL_DB_PATH")
	if path == "" {
		path = "./ritual.db"
	}

	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	cnxn, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := cnxn.Exec(schema); err != nil {
		return nil, err
	}

	DB = cnxn
	return DB, nil
}

func Close(db *sqlx.DB) error {
	if err := db.Close(); err != nil {
		return err
	}
	return nil
}
