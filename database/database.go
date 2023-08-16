package database

import (
	"database/sql"
	"fmt"
)

var db *sql.DB

const (
	host   = "localhost"
	port   = 5432 // Default port
	user   = "postgres"
	dbname = "spotlas-work"
)

func Connect() error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

func Get() *sql.DB {
	return db
}
