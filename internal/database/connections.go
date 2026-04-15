package database

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

const pgDSN = "host=localhost port=5432 user=postgres password=postgres dbname=url_shortener sslmode=disable"

func ConnectionPgSQL() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", pgDSN)
	if err != nil {
		log.Fatal("Ошибка при подключении к PostgreSQL: ", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}
