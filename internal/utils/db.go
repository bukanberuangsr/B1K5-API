package utils

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connectionStr := "host=postgres user=postgres password=secret123 dbname=b1k5_api port=5432 sslmode=disable"

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	DB = db
}
